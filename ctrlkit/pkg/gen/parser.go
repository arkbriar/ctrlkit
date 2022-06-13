package gen

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	errRedeclaration     = errors.New("redeclaration")
	errBindNotFound      = errors.New("bind not found")
	errTypeNotFound      = errors.New("type not found")
	errInvalidDeclBlock  = errors.New("invalid decl block")
	errInvalidStateBlock = errors.New("invalid state block")
	errInvalidActionStmt = errors.New("invalid action block")
)

func parseGvk(gvk string) (schema.GroupVersionKind, error) {
	lastSlash := strings.LastIndex(gvk, "/")
	if lastSlash < 0 || lastSlash == len(gvk)-1 {
		return schema.GroupVersionKind{}, errors.New("invalid gvk: " + gvk)
	}
	gv, err := parseGv(gvk[:lastSlash])
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	return schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    gvk[lastSlash+1:],
	}, nil
}

func parseGv(gv string) (schema.GroupVersion, error) {
	splits := strings.Split(gv, "/")
	if len(splits) == 1 && splits[0] == "v1" {
		return schema.GroupVersion{
			Group:   "",
			Version: "v1",
		}, nil
	}
	// At least two parts, {group/version}
	if len(splits) < 2 {
		return schema.GroupVersion{}, errors.New("invalid gv")
	}
	return schema.GroupVersion{
		Group:   strings.Join(splits[:len(splits)-1], "/"),
		Version: splits[len(splits)-1],
	}, nil
}

func splitLineIntoWords(line string) []string {
	wordScanner := bufio.NewScanner(bytes.NewBufferString(line))
	wordScanner.Split(bufio.ScanWords)

	words := make([]string, 0, 4)
	for wordScanner.Scan() {
		words = append(words, wordScanner.Text())
	}
	return words
}

func parseBind(words []string) (gv string, pkg string, gvParsed schema.GroupVersion, err error) {
	if len(words) != 3 {
		return "", "", schema.GroupVersion{}, errors.New("words size not match")
	}

	gv, pkg = words[1], words[2]
	gvParsed, err = parseGv(gv)

	return
}

func parseAlias(words []string) (gvk string, alias string, err error) {
	if len(words) != 3 {
		return "", "", errors.New("words size not match")
	}

	if strings.Contains(words[1], "/") {
		return "", "", errors.New("contains slash")
	}

	return words[2], words[1], nil
}

func parseDeclOpen(words []string) (name string, targetType string, err error) {
	if len(words) != 5 {
		return "", "", errors.New("words size not match")
	}
	if words[2] != "for" {
		return "", "", errInvalidDeclBlock
	}
	if !isBeginBracket(words[4:]) {
		return "", "", errInvalidDeclBlock
	}

	name, targetType = words[1], words[3]

	return
}

func isBeginBracket(words []string) bool {
	return len(words) == 1 && words[0] == "{"
}

func isEndBracket(words []string) bool {
	return len(words) == 1 && words[0] == "}"
}

// ParseDoc parses the bytes from reader into a structured ControllerManagerDocument
// when possible.
func ParseDoc(r io.Reader) (*ControllerManagerDocument, error) {
	doc := &ControllerManagerDocument{
		GvReflections: GvReflections{
			GvPkgBinds: make(map[string]GvBind),
			GvkAliases: make(map[string]string),
		},
		Decls: make(map[string]ControllerManagerDeclaration),
	}

	scanner := bufio.NewScanner(r)

	var comments []string
	var lastIsEmpty bool
	var lineNo int
	var inDecl, inState, inActions, inStateDecl bool
	var decl *ControllerManagerDeclaration
	var stateDecl *StateDeclaration

	// Parse the document line by line.
	for scanner.Scan() {
		line := scanner.Text()
		lineNo++

		// Trim leading and tail spaces.
		line = strings.TrimSpace(line)

		// Skip empty lines.
		if line == "" {
			lastIsEmpty = true
			continue
		}

		// Do parse.
		words := splitLineIntoWords(line)

		// Handle comments first. Keep the last continuous block of comments.
		if strings.HasPrefix(words[0], "//") {
			if lastIsEmpty {
				comments = nil
				lastIsEmpty = false
			}
			doubleSlashIndex := strings.Index(line, "//")
			comments = append(comments, strings.TrimSpace(line[doubleSlashIndex+2:]))
			continue
		}

		if inDecl {
			if inState {
				if inStateDecl {
					if isEndBracket(words) {
						decl.AddStateDeclaration(*stateDecl)

						comments = nil
						stateDecl = nil
						inStateDecl = false
					} else {
						if len(words) != 1 {
							return nil, fmt.Errorf("parse error: invalid state block at line %d: %w", lineNo, errors.New("size not match"))
						}
						splits := strings.Split(words[0], "=")
						if len(splits) > 2 {
							return nil, fmt.Errorf("parse error: invalid state block at line %d: %w", lineNo, errors.New("invalid selector"))
						} else if len(splits) == 2 {
							stateDecl.AddSelector(strings.TrimSpace(splits[0]), strings.TrimSpace(splits[1]))
						} else {
							stateDecl.AddSelector(strings.TrimSpace(splits[0]), "")
						}
					}
				} else {
					if isEndBracket(words) {
						comments = nil
						inState = false
					} else {
						if len(words) != 3 || !isBeginBracket(words[2:]) {
							return nil, fmt.Errorf("parse error: invalid state block at line %d: %w", lineNo, errInvalidStateBlock)
						}

						stateType := words[1]
						isArray := strings.HasPrefix(stateType, "[]")
						if isArray {
							stateType = stateType[2:]
						}
						if !doc.DoesAliasExists(stateType) {
							return nil, fmt.Errorf("parse error: invalid state block at line %d: %w", lineNo, errTypeNotFound)
						}

						// TODO check if state type exists

						stateDecl = &StateDeclaration{
							Comments:  comments,
							Name:      words[0],
							Type:      stateType,
							IsArray:   isArray,
							Selectors: make(map[string]string),
						}

						comments = nil
						inStateDecl = true
					}
				}
			} else if inActions {
				if isEndBracket(words) {
					comments = nil
					inActions = false
				} else {
					if line[0] == '(' || !strings.Contains(line, "(") || line[len(line)-1] != ')' {
						return nil, fmt.Errorf("parse error: invalid action declaration at line %d", lineNo)
					}
					splits := strings.SplitN(line[:len(line)-1], "(", 2)
					if strings.ContainsAny(splits[1], "()") {
						return nil, fmt.Errorf("parse error: invalid action declaration at line %d", lineNo)
					}
					name, params := splits[0], lo.Map(strings.Split(splits[1], ","), func(s string, i int) string {
						return strings.TrimSpace(s)
					})
					if len(params) == 1 && params[0] == "" {
						params = nil
					} else {
						for _, param := range params {
							if !decl.ContainsState(param) {
								return nil, fmt.Errorf("parse error: invalid action declaration at line %d: %w", lineNo, errTypeNotFound)
							}
						}
					}
					if !decl.AddActionDeclaration(ActionDeclaration{
						Comments: comments,
						Name:     name,
						Params:   params,
					}) {
						return nil, fmt.Errorf("parse error: invalid action declaration at line %d: %w", lineNo, errRedeclaration)
					}
					comments = nil
				}
			} else {
				if isEndBracket(words) {
					doc.Decls[decl.Name] = *decl

					comments = nil
					inDecl = false
					decl = nil
				} else {
					switch words[0] {
					case "state":
						if !isBeginBracket(words[1:]) {
							return nil, fmt.Errorf("parse error: invalid decl statement at line %d: %w", lineNo, errInvalidStateBlock)
						}
						comments = nil
						inState = true
					case "action":
						if !isBeginBracket(words[1:]) {
							return nil, fmt.Errorf("parse error: invalid decl statement at line %d: %w", lineNo, errInvalidActionStmt)
						}
						comments = nil
						inActions = true
					default:
						return nil, fmt.Errorf("parse error: invalid decl statement at line %d", lineNo)
					}
				}
			}
		} else {
			switch words[0] {
			case "bind":
				gv, pkg, gvParsed, err := parseBind(words)
				if err != nil {
					return nil, fmt.Errorf("parse error: invalid bind statement at line %d: %w", lineNo, err)
				}
				if !doc.AddGvBind(gv, pkg, gvParsed) {
					return nil, fmt.Errorf("parse error: invalid bind statement at line %d: %w", lineNo, errRedeclaration)
				}
				comments = nil
			case "alias":
				gvk, alias, err := parseAlias(words)
				if err != nil {
					return nil, fmt.Errorf("parse error: invalid alias statement at line %d: %w", lineNo, err)
				}
				gvkParsed, err := parseGvk(gvk)
				if err != nil {
					return nil, fmt.Errorf("parse error: invalid alias statement at line %d: %w", lineNo, err)
				}
				if !doc.IsGvBound(gvkParsed.GroupVersion().String()) {
					return nil, fmt.Errorf("parse error: invalid alias statement at line %d: %w", lineNo, errTypeNotFound)
				}
				if !doc.AddGvkAliases(gvk, alias) {
					return nil, fmt.Errorf("parse error: invalid alias statement at line %d: %w", lineNo, errRedeclaration)
				}
				comments = nil
			case "decl":
				name, targetType, err := parseDeclOpen(words)
				if err != nil {
					return nil, fmt.Errorf("parse error: invalid decl statement at line %d: %w", lineNo, err)
				}
				if doc.DoesControllerManagerDeclarationExists(name) {
					return nil, fmt.Errorf("parse error: invalid decl statement at line %d: %w", lineNo, errRedeclaration)
				}
				if !doc.IsGvBound(targetType) && !doc.DoesAliasExists(targetType) {
					return nil, fmt.Errorf("parse error: invalid decl statement at line %d: %w", lineNo, errBindNotFound)
				}

				inDecl = true
				decl = &ControllerManagerDeclaration{
					Comments:   comments,
					Name:       name,
					TargetType: targetType,
					States:     make(map[string]StateDeclaration),
					Actions:    nil,
					ActionMap:  make(map[string]ActionDeclaration),
				}
				// decl.AddStateDeclaration(StateDeclaration{
				// 	Comments: nil,
				// 	Name:     "target",
				// 	Type:     targetType,
				// })
				comments = nil
			default:
				return nil, fmt.Errorf("parse error: invalid statement at line %d", lineNo)
			}
		}

		lastIsEmpty = false
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return doc, nil
}
