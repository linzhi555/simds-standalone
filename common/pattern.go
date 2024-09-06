package common

// TODO this matchPatern is fake patern match,only support the ,need someone to fix!
func MatchPattern(p string, s string) bool {
	sLen := len(s)
	pLen := len(p)

	if p == "" {
		return p == s
	}

	// optimization for this project.in most case the only star is in the end of the pattern
	if p[pLen-1] == '*' &&
		sLen >= pLen-1 &&
		MatchPattern(p[0:pLen-1], s[0:pLen-1]) {
		return true
	}

	state := 0 // state can be [0,pLen] , pLen is receive state.
	for i := 0; i < sLen; i++ {
		if state == pLen {
			return false
		}
		switch p[state] {
		case '*':
			if MatchPattern(p[state+1:], s[i:]) {
				return true
			}
			if i == sLen-1 {
				if MatchPattern(p[state+1:], s[i+1:]) {
					return true
				}
			}

		default:
			if p[state] == s[i] {
				state++
			}else{
				return false
			}
		}

	}
	return state == pLen
}
