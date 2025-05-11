package engine

import (
	"regexp"
	"strings"
)

type Block struct {
	Type    string
	Phase   string
	Code    string
	LineNum int
}

const (
	Initial = "initial"
	Reboot  = "reboot"
	Reload  = "reload"
	Image   = "image"
)

var (
	codeBlockRe = regexp.MustCompile(`^([a-zA-Z]+)\s*(\{([^}]+)\})?$`)
)

func Parse(data string) (blocks []*Block, err error) {
	blocks = []*Block{}

	var curBlock *Block

	for n, line := range strings.Split(data, "\n") {
		if curBlock == nil {
			if strings.HasPrefix(line, "```") {
				lang, attrs := parseCodeBlockHeader(line[3:])

				phase := Initial
				if attrs != nil {
					if attrs["phase"] == Reboot {
						phase = Reboot
					} else if attrs["phase"] == Reload {
						phase = Reload
					}
				}

				switch lang {
				case "shell":
					curBlock = &Block{
						Type:  "shell",
						Phase: phase,
					}
				case "python":
					curBlock = &Block{
						Type:  "python",
						Phase: phase,
					}
				}
			}
		} else {
			if line == "```" {
				curBlock.LineNum = n + 1
				blocks = append(blocks, curBlock)
				curBlock = nil
			} else {
				curBlock.Code += line + "\n"
			}
		}
	}

	return
}

func parseCodeBlockHeader(input string) (language string,
	attrs map[string]string) {

	attrs = map[string]string{}

	matches := codeBlockRe.FindStringSubmatch(input)
	if len(matches) == 0 {
		return
	}

	language = matches[1]
	if len(matches) < 3 {
		return
	}

	attrPairs := strings.Split(matches[2], ",")
	for _, pair := range attrPairs {
		pair = strings.TrimPrefix(pair, "{")
		pair = strings.TrimSuffix(pair, "}")

		keyValue := strings.SplitN(pair, "=", 2)
		if len(keyValue) == 2 {
			key := strings.TrimSpace(keyValue[0])
			value := strings.Trim(strings.TrimSpace(keyValue[1]), `"`)
			attrs[key] = value
		}
	}

	return
}
