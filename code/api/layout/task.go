package layout

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type variable struct {
	static bool
	value  string
}

type op byte

const (
	opLess op = iota
	opGreater
	opEqual
	opNotEqual
	opLessEqual
	opGreaterEqual
)

func (op op) Run(left, right variable, args map[string]string) bool {
	var leftString, rightString string
	if left.static {
		leftString = left.value
	} else {
		leftString = args[left.value]
	}
	if right.static {
		rightString = right.value
	} else {
		rightString = args[right.value]
	}
	isNumber := true
	leftInt, err := strconv.ParseInt(leftString, 10, 64)
	if err != nil {
		isNumber = false
	}
	rightInt, err := strconv.ParseInt(rightString, 10, 64)
	if err != nil {
		isNumber = false
	}
	switch op {
	case opEqual:
		if isNumber {
			return leftInt == rightInt
		}
		return leftString == rightString
	case opNotEqual:
		if isNumber {
			return leftInt != rightInt
		}
		return leftString != rightString
	case opLess:
		if isNumber {
			return leftInt < rightInt
		}
		return leftString < rightString
	case opGreater:
		if isNumber {
			return leftInt > rightInt
		}
		return leftString > rightString
	case opLessEqual:
		if isNumber {
			return leftInt <= rightInt
		}
		return leftString <= rightString
	case opGreaterEqual:
		if isNumber {
			return leftInt >= rightInt
		}
		return leftString >= rightString
	default:
		return false
	}
}

type taskInfo struct {
	parent  *runner
	name    string
	auth    string
	timeout time.Duration
	hasIf   bool
	left    variable
	right   variable
	op      op
}

func isOP(str string) bool {
	if len(str) != 1 {
		return false
	}
	return str[0] == '>' ||
		str[0] == '<' ||
		str[0] == '!' ||
		str[0] == '='
}

func (info *taskInfo) Name() string {
	return info.name
}

func (info *taskInfo) parseIf(str string) error {
	info.hasIf = true
	str = strings.TrimSpace(str)
	var tmp []string
	var buf string
	for _, ch := range str {
		if isOP(string(ch)) {
			if len(buf) > 0 {
				tmp = append(tmp, strings.TrimSpace(buf))
			}
			tmp = append(tmp, string(ch))
			buf = ""
			continue
		}
		buf += string(ch)
	}
	if len(buf) > 0 {
		tmp = append(tmp, strings.TrimSpace(buf))
	}
	trans := make([]string, 0, len(tmp))
	var i int
	for {
		if i >= len(tmp) {
			break
		}
		if isOP(tmp[i]) {
			str := tmp[i]
			for j := i + 1; j < len(tmp); j++ {
				if !isOP(tmp[j]) {
					i = j
					break
				}
				str += tmp[j]
			}
			trans = append(trans, strings.TrimSpace(str))
		} else {
			trans = append(trans, strings.TrimSpace(tmp[i]))
			i++
		}
	}
	if len(trans) != 3 {
		return fmt.Errorf("invalid if expression")
	}
	switch trans[1] {
	case "<":
		info.op = opLess
	case ">":
		info.op = opGreater
	case "=":
		info.op = opEqual
	case "!=":
		info.op = opNotEqual
	case "<=":
		info.op = opLessEqual
	case ">=":
		info.op = opGreaterEqual
	default:
		return fmt.Errorf("invalid operator")
	}
	if trans[0][0] == '$' {
		info.left.static = false
		info.left.value = trans[0][1:]
	} else {
		info.left.static = true
		info.left.value = trans[0]
	}
	if trans[2][0] == '$' {
		info.right.static = false
		info.right.value = trans[2][1:]
	} else {
		info.right.static = true
		info.right.value = trans[2]
	}
	return nil
}

func (info *taskInfo) deadline() time.Time {
	if info.timeout > 0 {
		return time.Now().Add(info.timeout)
	}
	return info.parent.deadline
}

func (info *taskInfo) WantRun(args map[string]string) bool {
	if !info.hasIf {
		return true
	}
	return info.op.Run(info.left, info.right, args)
}
