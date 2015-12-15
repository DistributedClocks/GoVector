package brokervec

// A FilterMessage is a message from a subscriber containing the filter it
// wishes to apply to a Message's field in the form of a regular expression.
type FilterMessage struct {
    Regex string
    Nonce string
}

func (fm *FilterMessage) GetFilter() string {
        
    return fm.Regex
}
func (fm *FilterMessage) GetNonce() string {
        
    return fm.Nonce
}