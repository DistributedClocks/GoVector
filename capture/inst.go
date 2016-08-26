package capture

import(

	"fmt"
	"regexp"
	"sort"
	"bitbucket.org/bestchai/dinv/programslicer"

	"go/ast"
	"go/types"
	"go/token"
	"go/printer"
	"bytes"
	"golang.org/x/tools/go/ast/astutil"
	"go/importer"
)

var (
	debug         = false
	Directory = ""
	File      = ""
	Pipe		  = ""
)

//types of communication calls
const (
	SENDING = iota
	RECEIVING
	BOTH
	NOT
)


func InsturmentComm (options map[string]string) map[string]string {
	initalize(options)
	p, err := getProgramWrapper()
	if err != nil {
		panic(err)
	}
	netConns := GetNetConns(p)
	instrumentedOutput := make(map[string]string)
	for pnum, pack := range p.Packages {
		for snum, _ := range pack.Sources {
			InstrumentCalls(p,pnum,snum, netConns)
			buf := new(bytes.Buffer)
			printer.Fprint(buf, p.Fset, p.Packages[pnum].Sources[snum].Comments)
			p.Packages[pnum].Sources[snum].Text = buf.String()
			instrumentedOutput[p.Packages[pnum].Sources[snum].Filename] = buf.String()
		}
	}
	return instrumentedOutput
}



func initalize(options map[string]string) {
for setting := range options {
		switch setting {
		case "debug":
			debug = true
		case "directory":
			Directory = options[setting]
		case "file":
			File = options[setting]
		case "pipe":
			Pipe = options[setting]
		default:
			continue
		}
	}
}

func getProgramWrapper() (*programslicer.ProgramWrapper, error) {
	var (
		program *programslicer.ProgramWrapper
		err     error
	)
	if Directory != "" {
		program, err = programslicer.GetProgramWrapperDirectory(Directory)
		if err != nil {
			return program, err
		}
	} else if File != "" {
		program, err = programslicer.GetProgramWrapperFile(File)
		if err != nil {
			return program, err
		}
	} else if Pipe != "" {
		program, err = programslicer.GetWrapperFromString(Pipe)
		if err != nil {
			return program, err
		}
	}
	return program, nil
}

func GetCommNodes(p *programslicer.ProgramWrapper) (sending, receiving, both []*ast.Node){
	netConns := GetNetConns(p)
	for _, pack := range p.Packages {
		for _, src := range pack.Sources {
			ast.Inspect(src.Source, func(n ast.Node) bool {
				status := NOT // the type of call that was made
				switch z := n.(type) {
				case *ast.ExprStmt:
					switch c := z.X.(type) {
					case *ast.CallExpr:
						status = checkCall(p,c,netConns)
					}
				case *ast.AssignStmt:
					switch c := z.Rhs[0].(type) {
					case *ast.CallExpr:
						status = checkCall(p,c,netConns) 
					}
				}
				switch status {
				case SENDING:
					sending = append(sending,&n)
					break
				case RECEIVING:
					receiving = append(receiving,&n)
					break
				case BOTH:
					both = append(both,&n)
					break
				default:
					break
				}
			return true
			})
		}
	}
	fmt.Printf("sending %d receiving %d both %d",len(sending),len(receiving),len(both))
	return
}

//check call takes program wrapper, call expersion and set of
//POPULATED net connections as an arguments. The call expression is
//checked againts the set of known net connections for type. If the
//an enum denoting the type of connection is returned
func checkCall(p *programslicer.ProgramWrapper, c *ast.CallExpr, netConns map[types.Object]*NetConn) int {
	switch f := c.Fun.(type){
	case *ast.SelectorExpr:
		switch i:= f.X.(type){
		case *ast.Ident:
			for obj, conn := range netConns {
				if (i.Obj != nil) {
					//fmt.Printf("obj-Id: %s\t obj-Pkg: %s\tsearch-Obj.Name: %s\t searchObjPos:%d\t obj-Pos%d\n",obj.Id(),obj.Pkg().Name(),i.Name,i.Obj.Pos(),obj.Pos())
				}
				if (i.Obj != nil  && obj.Pos() == i.Obj.Pos()) || (obj.Pkg().Name() == i.Name) {
					fmt.Println("MATCH")
					if check(f.Sel.Name,conn.ReceivingFunctions,c,p){
						return RECEIVING
					} else if check(f.Sel.Name,conn.SenderFunctions,c,p){
						return SENDING
					} else if check(f.Sel.Name,conn.ConnectionFunctions,c,p) {
						return BOTH
					} else {
						return NOT
					}
				}
			}
		}
	}
	return NOT
}

//check compaires a variable name with the nown set of
//networking functions. If a match is made, true is returned.
func check(varName string, netfuncs []*NetFunc, call *ast.CallExpr, p *programslicer.ProgramWrapper) bool {
	for _, netFunc := range netfuncs {
		fmt.Printf("varName = %s, netFunc.Name = %s\n",varName,netFunc.Name)
		if varName == netFunc.Name {
			return true
		}
	}
	return false
}



func InstrumentCalls (p *programslicer.ProgramWrapper, pnum,snum int, netConns map[types.Object]*NetConn ){
	injected := false
	ast.Inspect(p.Packages[pnum].Sources[snum].Comments , func(n ast.Node) bool {
		switch c := n.(type){
		case *ast.CallExpr:
			switch f := c.Fun.(type){
			case *ast.SelectorExpr:
				switch i:= f.X.(type){
				case *ast.Ident:
					for obj, conn := range netConns {
						if (i.Obj != nil) {
							//fmt.Printf("nc-obj-Pkg: %s\t nc-obj-Id: %s\t\t found-obj-name: %s\t\t searchObjPos:%d\t\t obj-Pos:%d\n",obj.Pkg().Name(),obj.Id(), i.Name,i.Obj.Pos(),obj.Pos())
						}
						if (i.Obj != nil  && obj.Pos() == i.Obj.Pos()) || (obj.Pkg().Name() == i.Name) {
							//fmt.Println("MATCH")
							injected = checkAndInstrument(f.Sel.Name,conn.ReceivingFunctions,c,p) || injected
							injected = checkAndInstrument(f.Sel.Name,conn.SenderFunctions,c,p) || injected
							injected = checkAndInstrument(f.Sel.Name,conn.ConnectionFunctions,c,p) || injected
						}
					}
				}
			}
		}
		return true
	})
	//if code was added, add the apropriate import
	if injected {
		astutil.AddImport(p.Fset,p.Packages[pnum].Sources[snum].Comments,"github.com/arcaneiceman/GoVector/capture")
	}

}


//checkAndInstrument compaires a variable name with the nown set of
//networking functions. If the variable is found to be a networking
//function, it is captured.
//If the variable is insturmented the function returns true.
func checkAndInstrument(varName string, netfuncs []*NetFunc, call *ast.CallExpr, p *programslicer.ProgramWrapper) bool {
	for _, netFunc := range netfuncs {
		//fmt.Printf("varName = %s, netFunc.Name = %s\n",varName,netFunc.Name)
		if varName == netFunc.Name {
			instrumentCall(call,netFunc)
			return true
		}
	}
	return false
}


func instrumentCall(call *ast.CallExpr, function *NetFunc){
	funString := getFunctionString(call)
	args := getArgsString(call.Args)
	call.Fun = ast.NewIdent(fmt.Sprintf("capture.%s",function.Name))
	call.Args = make([]ast.Expr,1)
	call.Args[0] = ast.NewIdent(fmt.Sprintf("%s,%s",funString,args))
}

func getFunctionString(call *ast.CallExpr) string {
	fset := token.NewFileSet()
	var buf bytes.Buffer
	printer.Fprint(&buf,fset,call.Fun)
	return buf.String()
}

func getArgsString(args []ast.Expr)string {
	fset := token.NewFileSet()
	var buf bytes.Buffer
	argsString := make([]string,0)
	for i:= range args {
		buf.Reset()
		printer.Fprint(&buf,fset,args[i])
		argsString = append(argsString,buf.String())
	}

	var output string
	for i:= 0;i < len(argsString) -1;i++{
		output += argsString[i] + ","
	}
	output += argsString[len(argsString) -1]
	return output
}


//TODO merge with Get SendReceiveNodes
//GetNetConns searches through a program for net connections, and
//adds their object reference to a known database
func GetNetConns(program *programslicer.ProgramWrapper) map[types.Object]*NetConn {
	// Type-check the package.
    // We create an empty map for each kind of input
    // we're interested in, and Check populates them.
    info := types.Info{
            Types: make(map[ast.Expr]types.TypeAndValue),
            Defs:  make(map[*ast.Ident]types.Object),
            Uses:  make(map[*ast.Ident]types.Object),
    }
	//var conf types.Config
	conf := types.Config{Importer: importer.Default()}

	sources := make([]*ast.File,0)
	for _, source := range program.Packages[0].Sources {	//TODO extend to interpackage
		sources = append(sources,source.Comments) //NOTE CHANGED FROM SOURCES (SHOULD BE SOURCES) /Trying to figure out why obj-pos() not working for the checking function
	}
    pkg, err := conf.Check(program.Packages[0].PackageName, program.Fset, sources, &info) //TODO extend to interpackage
    if err != nil {
            fmt.Println(err)
    }

    usesByObj := make(map[types.Object][]string)
    for id, obj := range info.Uses {
            posn := program.Fset.Position(id.Pos())
            lineCol := fmt.Sprintf("%d:%d", posn.Line, posn.Column)
            usesByObj[obj] = append(usesByObj[obj], lineCol)
    }
	//capture variables name type
	revar := regexp.MustCompile(`var ([A-Za-z0-9_]+) \**([*.A-Za-z0-9_]+)`)
	refunc := regexp.MustCompile(`func ([A-Za-z0-9_/]+).([A-Za-z0-9_]+)\(`)
	var netConns map[types.Object]*NetConn = make(map[types.Object]*NetConn,0)


    for obj, uses := range usesByObj {
		sort.Strings(uses)
		ObjectDef := types.ObjectString(obj, types.RelativeTo(pkg))

		//fmt.Println(ObjectDef)
		//variables
		match := revar.FindStringSubmatch(ObjectDef)
		//fmt.Println(match)
		if len(match) == 3 {
			//type is in netvars
			_, ok := NetDB[match[2]]	//match[2] contians the type of the object
			if ok {
				netConns[obj] = NetDB[match[2]]
			}
		}
		//variables
		match = refunc.FindStringSubmatch(ObjectDef)
		//fmt.Println(match)
		if len(match) == 3 {
			//type is in netvars
			_, ok := NetDB[match[1]]	//match[2] contians the type of the object
			if ok {
				netConns[obj] = NetDB[match[1]]
				//fmt.Printf("Added Obj: %s to the DB\n",obj.String())
			}
		} 
	}
	for conn := range netConns {
		fmt.Println(conn.String())
	}

	return netConns
	//NOTE This is where the differences between detects
	//GetSendReceiveNodes differes
}

type NetConn struct {
	NetType string
	SenderFunctions []*NetFunc
	ReceivingFunctions []*NetFunc
	ConnectionFunctions []*NetFunc
}

type NetFunc struct {
	Name string
	Args int
	Returns int
	PrimaryArgLoc int
	ReturnSizeLoc int
}

var NetDB map[string]*NetConn = map[string]*NetConn{
	"net.UDPConn" : &NetConn{
						"net.UDPConn",
						[]*NetFunc{
							&NetFunc{"Write",1,2,0,0},
							&NetFunc{"WriteMsgUDP",2,2,0,0},
							&NetFunc{"WriteTo",2,2,0,0},
							&NetFunc{"WriteToUDP",2,2,0,0},
						},
						[]*NetFunc{
							&NetFunc{"Read",1,2,0,0},
							&NetFunc{"ReadFrom",1,3,0,0},
							&NetFunc{"ReadFromUDP",1,3,0,0},
							&NetFunc{"ReadMsgUDP",2,5,0,0},
						},
						[]*NetFunc{},
					},
	"net.UnixConn" : &NetConn{
						"net.UnixConn",
						[]*NetFunc{
							&NetFunc{"Write",1,2,0,0},
							&NetFunc{"WriteMsgUnix",3,3,0,0},
							&NetFunc{"WriteTo",2,2,0,0},
							&NetFunc{"WriteToUnix",2,2,0,0},
						},
						[]*NetFunc{
							&NetFunc{"Read",1,2,0,0},
							&NetFunc{"ReadFrom",1,3,0,0},
							&NetFunc{"ReadFromUDP",1,3,0,0},
							&NetFunc{"ReadMsgUDP",2,5,0,0},
						},
						[]*NetFunc{},
					},
	"net.IPConn" : &NetConn{
						"net.IPConn",
						[]*NetFunc{
							&NetFunc{"Write",1,2,0,0},
							&NetFunc{"WriteMsgIP",2,2,0,0},
							&NetFunc{"WriteTo",2,2,0,0},
							&NetFunc{"WriteToIP",2,2,0,0},
						},
						[]*NetFunc{
							&NetFunc{"Read",1,2,0,0},
							&NetFunc{"ReadFrom",1,3,0,0},
							&NetFunc{"ReadFromIP",1,3,0,0},
							&NetFunc{"ReadMsgIP",2,5,0,0},
						},
						[]*NetFunc{},
					},
	"net.TCPConn" : &NetConn{
						"net.TCPConn",
						[]*NetFunc{
							&NetFunc{"Write",1,2,0,0},
						},
						[]*NetFunc{
							&NetFunc{"Read",1,2,0,0},
							//&NetFunc{"ReadFrom",1,3,0,0}, //This
							//weird method takes an io.Reader as an
							//argument rather that a buffer
						},
						[]*NetFunc{},
					},
	"net.PacketConn" : &NetConn{
						"net.PacketConn",
						[]*NetFunc{
							&NetFunc{"WriteTo",2,2,0,0},
						},
						[]*NetFunc{
							&NetFunc{"ReadFrom",1,3,0,0},
						},
						[]*NetFunc{},
					},
	"net.Conn" : &NetConn{
						"net.Conn",
						[]*NetFunc{
							&NetFunc{"Write",2,2,0,0},
						},
						[]*NetFunc{
							&NetFunc{"Read",1,3,0,0},
						},
						[]*NetFunc{},
					},
					/*
	"net/rpc.Server" : &NetConn{
						"rpc.Server",
						[]*NetFunc{},
						[]*NetFunc{},
						[]*NetFunc{
							&NetFunc{"ServerCodec",1,0,0,0},
							&NetFunc{"ServerConn",1,0,0,0},
							&NetFunc{"ServeHTTP",2,0,?,0},
							&NetFunc{"ServeRequest",1,0,0,0},
						},
					},*/
	"net/rpc" : &NetConn{
						"rpc",
						[]*NetFunc{},
						[]*NetFunc{},
						[]*NetFunc{
							&NetFunc{"Dial",2,2,0,0},
							&NetFunc{"DialHTTP",3,2,0,0},
							&NetFunc{"DialHTTPPath",3,2,0,0},
							&NetFunc{"NewClient",1,0,0,0},
							&NetFunc{"NewClientWithCodec",1,0,0,0},
							&NetFunc{"ServeCodec",1,0,0,0},
							&NetFunc{"ServeConn",1,0,0,0},
							&NetFunc{"ServeRequest",1,0,0,0},
						},
					},
				}

