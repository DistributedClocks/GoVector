package assert

// Constants of operations
const (
	opDEL byte = iota
	opINS
	opCHANGE
	opMATCH
)

func matchingFromOps(la, lb int, ops []byte) (matA, matB []int) {
	matA, matB = make([]int, la), make([]int, lb)
	for i, j := la, lb; i > 0 || j > 0; {
		var op byte
		switch {
		case i == 0:
			op = opINS
		case j == 0:
			op = opDEL
		default:
			op = ops[(i-1)*lb+j-1]
		}

		switch op {
		case opINS:
			j--
			matB[j] = -1
		case opDEL:
			i--
			matA[i] = -1
		case opCHANGE:
			i--
			j--
			matA[i], matB[j] = j, i
		}
	}

	return matA, matB
}

func match(lenA, lenB int, costOfChange func(iA, iB int) int, costOfDel func(iA int) int, costOfIns func(iB int) int) (dist int, matA, matB []int) {
	la, lb := lenA, lenB

	f := make([]int, lb+1)
	ops := make([]byte, la*lb)

	for j := 1; j <= lb; j++ {
		f[j] = f[j-1] + costOfIns(j-1)
	}

	// Matching with dynamic programming
	p := 0
	for i := 0; i < la; i++ {
		fj1 := f[0] // fj1 is the value of f[j - 1] in last iteration
		f[0] += costOfDel(i)
		for j := 1; j <= lb; j++ {
			mn, op := f[j]+costOfDel(i), opDEL // delete

			if v := f[j-1] + costOfIns(j-1); v < mn {
				// insert
				mn, op = v, opINS
			}

			// change/matched
			if v := fj1 + costOfChange(i, j-1); v < mn {
				// insert
				mn, op = v, opCHANGE
			}

			fj1, f[j], ops[p] = f[j], mn, op // save f[j] to fj1(j is about to increase), update f[j] to mn
			p++
		}
	}
	// Reversely find the match info
	matA, matB = matchingFromOps(la, lb, ops)

	return f[lb], matA, matB
}
