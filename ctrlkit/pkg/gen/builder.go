package gen

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samber/lo"
)

var CtrlKitPackage = "github.com/arkbriar/ctrlkit/pkg/ctrlkit"

const goFileTemplate = `package %s

import (
	%s
)

%s
`

func formatIntoGoFile(pkg string, imports []string, body string) (string, error) {
	s := fmt.Sprintf(goFileTemplate, pkg, strings.Join(imports, "\n\t"), body)
	return s, nil
}

func constructPkgAliasForGvPkg(bind GvBind) string {
	if bind.Gv == "v1" {
		return "corev1"
	} else {
		slices := strings.Split(bind.Pkg, "/")
		if len(slices) < 2 {
			panic("impossible package for gvk")
		}
		if slices[len(slices)-1] == bind.Parsed.Version {
			return slices[len(slices)-2] + bind.Parsed.Version
		} else {
			return slices[len(slices)-1] + bind.Parsed.Version
		}
	}
}

func generateImports(doc *ControllerManagerDocument) ([]string, error) {
	// Constant imports.
	pkgMap := map[string]string{
		"context":                            "",
		"errors":                             "",
		"fmt":                                "",
		CtrlKitPackage:                       "",
		"k8s.io/apimachinery/pkg/api/errors": "apierrors",
		"k8s.io/apimachinery/pkg/types":      "",
		"sigs.k8s.io/controller-runtime/pkg/client": "",
		"sigs.k8s.io/controller-runtime":            "ctrl",
	}

	// Add each binds into the imports.
	for _, bind := range doc.GvPkgBinds {
		pkgMap[bind.Pkg] = constructPkgAliasForGvPkg(bind)
	}

	// Generate imports following the golang grammar.
	imports := make([]string, 0, len(pkgMap))
	for k, v := range pkgMap {
		if v != "" {
			imports = append(imports, fmt.Sprintf("%s \"%s\"", v, k))
		} else {
			imports = append(imports, "\""+k+"\"")
		}
	}

	return imports, nil
}

const managerStateGoTemplate = `// %sState is the state manager of %s.
type %sState struct {
	client.Reader
	target *%s
}
%s

// New%sState returns a %sState (target is not copied).
func New%sState(reader client.Reader, target *%s) %sState {
	return %sState{
		Reader: reader,
		target: target,
	}
}
`

func formatIntoStateGoCode(doc *ControllerManagerDocument, mgr *ControllerManagerDeclaration) (string, error) {
	bodyBuf := &bytes.Buffer{}

	stateNames := lo.Keys(mgr.States)
	sort.Strings(stateNames)

	for _, stateName := range stateNames {
		stateDecl := mgr.States[stateName]
		bodyBuf.WriteRune('\n')
		s, err := generateGrabStatePolyfillCodes(doc, mgr, &stateDecl)
		if err != nil {
			return "", err
		}
		bodyBuf.WriteString(s)
	}

	typeGvk := doc.GetGvkByAlias(mgr.TargetType)
	gvk, err := parseGvk(typeGvk)
	if err != nil {
		return "", err
	}
	typeBind := doc.GvPkgBinds[gvk.GroupVersion().String()]
	targetGoType := constructPkgAliasForGvPkg(typeBind) + "." + gvk.Kind

	return fmt.Sprintf(managerStateGoTemplate,
		mgr.Name, mgr.Name,
		mgr.Name,
		targetGoType,
		bodyBuf.String(),
		mgr.Name, mgr.Name,
		mgr.Name, targetGoType, mgr.Name, mgr.Name,
	), nil
}

func formatSelectorsIntoComments(selectors map[string]string) string {
	buf := &bytes.Buffer{}

	sortedKeys := lo.Keys(selectors)
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		v := selectors[k]
		buf.WriteString("//   + ")
		buf.WriteString(k)
		if v != "" {
			buf.WriteString("=")
			buf.WriteString(v)
		}
		buf.WriteString("\n")
	}

	s := buf.String()
	return s[:len(s)-1]
}

const (
	managerStateMethodGetTemplate = `// Get%s gets %s with name equals to %s.
func (s *%sState) Get%s(ctx context.Context) (*%s, error) {
	var %s %s

	err := s.Get(ctx, types.NamespacedName{
		Namespace: s.target.Namespace,
		Name: %s,
	}, &%s)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get state '%s': %%w", err)
	}
	%s

	return &%s, nil
}
`

	managerStateMethodGetByListTemplate = `// Get%s gets %s with the following selectors:
%s
func (s *%sState) Get%s(ctx context.Context) (*%s, error) {
	var %sList %sList

	matchingLabels := map[string]string{
%s
	}
	matchingFields := map[string]string{
%s
	}

	err := s.List(ctx, &%sList, client.InNamespace(s.target.Namespace), 
		client.MatchingLabels(matchingLabels), client.MatchingFields(matchingFields))
	if err != nil {
		return nil, fmt.Errorf("unable to get state '%s': %%w", err)
	}

	if len(%sList.Items) == 0 {
		return nil, nil
	}
	if len(%sList.Items) != 1 {
		return nil, fmt.Errorf("unable to get state '%s': more than 1 object found")
	}

	%s := %sList.Items[0]
	%s

	return &%s, nil
}
`

	managerStateMethodListTemplate = `// Get%s lists %s with the following selectors:
%s
func (s *%sState) Get%s(ctx context.Context) ([]%s, error) {
	var %sList %sList

	matchingLabels := map[string]string{
%s
	}
	matchingFields := map[string]string{
%s
	}

	err := s.List(ctx, &%sList, client.InNamespace(s.target.Namespace), 
		client.MatchingLabels(matchingLabels), client.MatchingFields(matchingFields))
	if err != nil {
		return nil, fmt.Errorf("unable to get state '%s': more than 1 object found")
	}

	var validated []%s
	for _, obj := range %sList.Items {
%s
	}

	return validated, nil
}
`
)

func lowerTheFirstCharInWord(s string) string {
	if s == "" {
		return ""
	}
	return string(bytes.ToLower([]byte{s[0]})) + s[1:]
}

func lowerTheFirstCharInSentences(s string) string {
	split := strings.SplitN(s, " ", 2)
	if len(split) == 1 {
		return lowerTheFirstCharInWord(split[0])
	} else {
		return lowerTheFirstCharInWord(split[0]) + split[1]
	}
}

func upperTheFirstCharInWord(s string) string {
	if s == "" {
		return ""
	}
	return string(bytes.ToUpper([]byte{s[0]})) + s[1:]
}

func getStrExpr(expr string, targetStub string) (string, error) {
	var exprs []string
	i := 0
	for i < len(expr) {
		if i < len(expr)-2 && expr[i:i+2] == "${" {
			// consume to "}"
			rightBracketIndex := strings.Index(expr[i+2:], "}")
			if rightBracketIndex < 0 {
				return "", errors.New("unclosed reference brackets")
			}
			targetAndUse := strings.SplitN(expr[i+2:i+2+rightBracketIndex], ".", 2)
			if len(targetAndUse) < 2 {
				exprs = append(exprs, targetStub)
			} else if targetAndUse[0] == "target" {
				exprs = append(exprs, targetStub+"."+targetAndUse[1])
			} else {
				return "", errors.New("reference non-target not supported")
			}
			i += rightBracketIndex + 3
		} else if expr[i] == '$' {
			return "", errors.New("'$' is not allowed")
		} else {
			// consume to "$"
			nextDollarIndex := strings.Index(expr[i:], "$")
			if nextDollarIndex < 0 {
				exprs = append(exprs, "\""+expr[i:]+"\"")
				i = len(expr)
			} else {
				exprs = append(exprs, "\""+expr[i:nextDollarIndex]+"\"")
				i += nextDollarIndex + 1
			}
		}
	}
	return strings.Join(exprs, " + "), nil
}

func getStateNameExpr(state *StateDeclaration) (string, error) {
	nameSelector := state.Selectors["name"]
	return getStrExpr(nameSelector, "s.target")
}

func generateMapKeyAndValuesInGo(kv map[string]string, indent string) string {
	if len(kv) == 0 {
		return ""
	}

	buf := &bytes.Buffer{}

	keys := lo.Keys(kv)
	sort.Strings(keys)

	for _, k := range keys {
		v := kv[k]
		buf.WriteString(indent)
		buf.WriteRune('"')
		buf.WriteString(k)
		buf.WriteRune('"')
		buf.WriteRune(':')
		buf.WriteRune('"')
		buf.WriteString(v)
		buf.WriteRune('"')
		buf.WriteRune(',')
		buf.WriteRune('\n')
	}

	s := buf.String()

	return s[:len(s)-1]
}

func generateMapKeyAndUnquotedValuesInGo(kv map[string]string, indent string) string {
	if len(kv) == 0 {
		return ""
	}

	buf := &bytes.Buffer{}

	keys := lo.Keys(kv)
	sort.Strings(keys)

	for _, k := range keys {
		v := kv[k]
		buf.WriteString(indent)
		buf.WriteRune('"')
		buf.WriteString(k)
		buf.WriteRune('"')
		buf.WriteRune(':')
		buf.WriteString(v)
		buf.WriteRune(',')
		buf.WriteRune('\n')
	}

	s := buf.String()

	return s[:len(s)-1]
}

func generateMatchingLabels(state *StateDeclaration, indent string) (string, error) {
	labels := make(map[string]string)
	for k, v := range state.Selectors {
		if strings.HasPrefix(k, "labels/") {
			value, err := getStrExpr(v, "s.target")
			if err != nil {
				return "", err
			}
			labels[k[7:]] = value
		}
	}
	return generateMapKeyAndUnquotedValuesInGo(labels, indent), nil
}

func generateMatchingFields(state *StateDeclaration, indent string) (string, error) {
	labels := make(map[string]string)
	for k, v := range state.Selectors {
		if strings.HasPrefix(k, "fields/") {
			value, err := getStrExpr(v, "s.target")
			if err != nil {
				return "", err
			}
			labels[k[7:]] = value
		}
	}
	return generateMapKeyAndUnquotedValuesInGo(labels, indent), nil
}

func indentStr(s string, indent string) string {
	return strings.Join(lo.Map(strings.Split(s, "\n"), func(s string, i int) string {
		return indent + s
	}), "\n")
}

func generateCheckOwnership(stateVar string, trueBlock, falseBlock string, indent string) string {
	trueBranch, falseBranch := trueBlock != "", falseBlock != ""
	if trueBranch && falseBranch {
		return fmt.Sprintf(`if ctrlkit.ValidateOwnership(&%s, s.target) {
%s
} else {
%s
}`, stateVar, indentStr(trueBlock, "\t"), indentStr(falseBlock, "\t"))
	} else if trueBranch {
		return fmt.Sprintf(`if ctrlkit.ValidateOwnership(&%s, s.target) {
%s
}`, stateVar, indentStr(trueBlock, "\t"))
	} else if falseBranch {
		return fmt.Sprintf(`if !ctrlkit.ValidateOwnership(&%s, s.target) {
%s
}`, stateVar, indentStr(falseBlock, "\t"))
	} else {
		return ""
	}
}

func generateGetStateCodes(doc *ControllerManagerDocument, mgr *ControllerManagerDeclaration, state *StateDeclaration) (string, error) {
	typeGvk := doc.GetGvkByAlias(state.Type)
	gvk, err := parseGvk(typeGvk)
	if err != nil {
		return "", err
	}
	typeBind := doc.GvPkgBinds[gvk.GroupVersion().String()]
	stateGoType := constructPkgAliasForGvPkg(typeBind) + "." + gvk.Kind

	stateVarName := state.Name
	nameExpr, err := getStateNameExpr(state)
	if err != nil {
		return "", err
	}
	ownershipCheck := ""
	if _, ok := state.Selectors["owned"]; ok {
		ownershipCheck = generateCheckOwnership(stateVarName, "",
			fmt.Sprintf(`return nil, fmt.Errorf("unable to get state '%s': object not owned by target")`, state.Name), "\t")
		ownershipCheck = indentStr(ownershipCheck, "\t")
	}

	return fmt.Sprintf(managerStateMethodGetTemplate,
		upperTheFirstCharInWord(state.Name), state.Name, state.Selectors["name"],
		mgr.Name, upperTheFirstCharInWord(state.Name), stateGoType,
		stateVarName, stateGoType,
		nameExpr,
		stateVarName,
		state.Name,
		ownershipCheck,
		stateVarName,
	), nil
}

func generateGetStateByListCodes(doc *ControllerManagerDocument, mgr *ControllerManagerDeclaration, state *StateDeclaration) (string, error) {
	typeGvk := doc.GetGvkByAlias(state.Type)
	gvk, err := parseGvk(typeGvk)
	if err != nil {
		return "", err
	}
	typeBind := doc.GvPkgBinds[gvk.GroupVersion().String()]
	stateGoType := constructPkgAliasForGvPkg(typeBind) + "." + gvk.Kind

	stateVarName := state.Name

	matchingLabels, err := generateMatchingLabels(state, "\t\t")
	if err != nil {
		return "", err
	}
	matchingFields, err := generateMatchingFields(state, "\t\t")
	if err != nil {
		return "", err
	}

	ownershipCheck := ""
	if _, ok := state.Selectors["owned"]; ok {
		ownershipCheck = generateCheckOwnership(stateVarName, "",
			fmt.Sprintf(`return nil, fmt.Errorf("unable to get state '%s': object not owned by target")`, state.Name), "\t")
		ownershipCheck = indentStr(ownershipCheck, "\t")
	}

	return fmt.Sprintf(managerStateMethodGetByListTemplate,
		upperTheFirstCharInWord(state.Name), state.Name,
		formatSelectorsIntoComments(state.Selectors),
		mgr.Name,
		upperTheFirstCharInWord(state.Name),
		stateGoType,
		stateVarName,
		stateGoType,
		matchingLabels,
		matchingFields,
		stateVarName,
		state.Name,
		stateVarName,
		stateVarName,
		state.Name,
		stateVarName, stateVarName,
		ownershipCheck,
		stateVarName,
	), nil
}

func generateListStateCodes(doc *ControllerManagerDocument, mgr *ControllerManagerDeclaration, state *StateDeclaration) (string, error) {
	typeGvk := doc.GetGvkByAlias(state.Type)
	gvk, err := parseGvk(typeGvk)
	if err != nil {
		return "", err
	}
	typeBind := doc.GvPkgBinds[gvk.GroupVersion().String()]
	stateGoType := constructPkgAliasForGvPkg(typeBind) + "." + gvk.Kind

	stateVarName := state.Name

	matchingLabels, err := generateMatchingLabels(state, "\t\t")
	if err != nil {
		return "", err
	}
	matchingFields, err := generateMatchingFields(state, "\t\t")
	if err != nil {
		return "", err
	}

	// FIXME: Ownership check is an optional but the template requires it.
	ownershipCheck := ""
	if _, ok := state.Selectors["owned"]; ok {
		ownershipCheck = generateCheckOwnership("obj", "validated = append(validated, obj)", "", "\t")
		ownershipCheck = indentStr(ownershipCheck, "\t\t")
	}

	return fmt.Sprintf(managerStateMethodListTemplate,
		upperTheFirstCharInWord(state.Name), state.Name,
		formatSelectorsIntoComments(state.Selectors),
		mgr.Name,
		upperTheFirstCharInWord(state.Name),
		stateGoType,
		stateVarName,
		stateGoType,
		matchingLabels,
		matchingFields,
		stateVarName,
		state.Name,
		stateGoType,
		stateVarName,
		ownershipCheck,
	), nil
}

func generateGrabStatePolyfillCodes(doc *ControllerManagerDocument, mgr *ControllerManagerDeclaration, state *StateDeclaration) (string, error) {
	if _, containName := state.Selectors["name"]; containName {
		if state.IsArray {
			return "", errors.New("state is an array but selectors contain \"name\"")
		}
		return generateGetStateCodes(doc, mgr, state)
	} else {
		if state.IsArray {
			return generateListStateCodes(doc, mgr, state)
		} else {
			return generateGetStateByListCodes(doc, mgr, state)
		}
	}
}

const (
	managerStubCodeTemplate = `// %sImpl declares the implementation interface for %s.
type %sImpl interface {
	ctrlkit.CrontollerManagerActionLifeCycleHook

%s
}

%s
type %s struct {
	state 	%sState
	impl 	%sImpl
	logger	logr.Logger
}

// WrapAction returns an action from manager.
func (m *%s) WrapAction(description string, f func(context.Context, logr.Logger) (ctrl.Result, error)) ctrlkit.ReconcileAction {
	return ctrlkit.WrapAction(description, func(ctx context.Context) (ctrl.Result, error) {
		logger := m.logger.WithValues("action", description)

		defer m.impl.AfterActionRun(description, ctx, logger)
		m.impl.BeforeActionRun(description, ctx, logger)
		return f(ctx, logger)
	})
}

%s

// New%s returns a new %s with given state and implementation.
func New%s(state %sState, impl %sImpl, logger logr.Logger) %s {
	return %s{
		state: 	state,
		impl: 	impl,
		logger: logger,
	}
}
`
)

func getParamRefType(doc *ControllerManagerDocument, mgr *ControllerManagerDeclaration, param string) (string, error) {
	stateDecl := mgr.States[param]

	typeGvk := doc.GetGvkByAlias(stateDecl.Type)
	gvk, err := parseGvk(typeGvk)
	if err != nil {
		return "", err
	}
	typeBind := doc.GvPkgBinds[gvk.GroupVersion().String()]
	stateGoType := constructPkgAliasForGvPkg(typeBind) + "." + gvk.Kind

	if stateDecl.IsArray {
		return "[]" + stateGoType, nil
	} else {
		return "*" + stateGoType, nil
	}
}

func generateImplInterfaceDecl(doc *ControllerManagerDocument, mgr *ControllerManagerDeclaration) (string, error) {
	methods := make([]string, 0, len(mgr.Actions))

	for _, act := range mgr.Actions {
		method := ""
		if len(act.Comments) > 0 {
			method += strings.Join(lo.Map(act.Comments, func(s string, _ int) string {
				return "\t// " + s
			}), "\n") + "\n"
		}
		params := make([]string, 0, len(act.Params)+1)
		params = append(params, "ctx context.Context", "logger logr.Logger")
		for _, param := range act.Params {
			paramType, err := getParamRefType(doc, mgr, param)
			if err != nil {
				return "", err
			}

			params = append(params, param+" "+paramType)
		}
		method += "\t" + fmt.Sprintf("%s(%s) (ctrl.Result, error)", act.Name, strings.Join(params, ", "))

		methods = append(methods, method)
	}

	return strings.Join(methods, "\n\n"), nil
}

const (
	mgrMethodTemplate = `%s
func (m *%s) %s() ctrlkit.ReconcileAction {
	return ctrlkit.WrapAction("%s", %s)
}
`
)

func generateMgrMethodBody(doc *ControllerManagerDocument, mgr *ControllerManagerDeclaration, act *ActionDeclaration) string {
	const errHandleCode = `if err != nil {
	return ctrlkit.RequeueIfError(err)
}`
	buf := bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("logger := m.logger.WithValues(\"action\", \"%s\")\n\n", act.Name))

	if len(act.Params) > 0 {
		buf.WriteString("// Get states.\n")

		for _, param := range act.Params {
			buf.WriteString(param)
			buf.WriteString(", err := ")
			buf.WriteString("m.state.Get")
			buf.WriteString(upperTheFirstCharInWord(param))
			buf.WriteString("(ctx)\n")
			buf.WriteString(errHandleCode)
			buf.WriteString("\n\n")
		}

	}

	buf.WriteString("// Invoke action.\n")
	buf.WriteString(fmt.Sprintf("defer m.impl.AfterActionRun(\"%s\", ctx, logger)\n", act.Name))
	buf.WriteString(fmt.Sprintf("m.impl.BeforeActionRun(\"%s\", ctx, logger)\n\n", act.Name))

	buf.WriteString("return m.impl.")
	buf.WriteString(act.Name)
	buf.WriteString("(ctx, logger")
	if len(act.Params) > 0 {
		buf.WriteString(", ")
		buf.WriteString(strings.Join(act.Params, ", "))
	}
	buf.WriteString(")")

	return fmt.Sprintf(`func (ctx context.Context) (ctrl.Result, error) {
%s
}`, indentStr(buf.String(), "\t\t"))
}

func generateMgrMethods(doc *ControllerManagerDocument, mgr *ControllerManagerDeclaration) (string, error) {
	methods := make([]string, 0, len(mgr.Actions))

	for _, act := range mgr.Actions {
		method := fmt.Sprintf(mgrMethodTemplate,
			// strings.Join(lo.Map(act.Comments, func(s string, _ int) string {
			// 	return "\t// " + s
			// }), "\n"),
			"// "+act.Name+" generates the action of \""+act.Name+"\".",
			mgr.Name, act.Name,
			act.Name,
			generateMgrMethodBody(doc, mgr, &act),
		)
		methods = append(methods, method)
	}

	return strings.Join(methods, "\n"), nil
}

func formatIntoManagerGoCode(doc *ControllerManagerDocument, mgr *ControllerManagerDeclaration) (string, error) {
	var implInterfaceDecl, mgrMethods string
	var err error
	if implInterfaceDecl, err = generateImplInterfaceDecl(doc, mgr); err != nil {
		return "", err
	}
	if mgrMethods, err = generateMgrMethods(doc, mgr); err != nil {
		return "", err
	}

	return fmt.Sprintf(managerStubCodeTemplate,
		mgr.Name, mgr.Name,
		mgr.Name,
		implInterfaceDecl,
		strings.Join(lo.Map(mgr.Comments, func(s string, i int) string {
			return "// " + s
		}), "\n"),
		mgr.Name,
		mgr.Name,
		mgr.Name,
		mgr.Name,
		mgrMethods,
		mgr.Name, mgr.Name,
		mgr.Name, mgr.Name, mgr.Name, mgr.Name,
		mgr.Name,
	), nil
}

func generateBody(doc *ControllerManagerDocument) (string, error) {
	bodyBuf := &bytes.Buffer{}

	mgrNames := lo.Keys(doc.Decls)
	sort.Strings(mgrNames)
	for _, mgrName := range mgrNames {
		mgr := doc.Decls[mgrName]
		// Stub codes for state.
		stateStubCodes, err := formatIntoStateGoCode(doc, &mgr)
		if err != nil {
			return "", err
		}
		bodyBuf.WriteString(stateStubCodes)
		bodyBuf.WriteRune('\n')

		// Stub codes for manager.
		managerStubCodes, err := formatIntoManagerGoCode(doc, &mgr)
		if err != nil {
			return "", err
		}
		bodyBuf.WriteString(managerStubCodes)
		bodyBuf.WriteRune('\n')
	}

	return bodyBuf.String(), nil
}

func GenerateStubCodes(doc *ControllerManagerDocument, pkgName string) (string, error) {
	imports, err := generateImports(doc)
	if err != nil {
		return "", err
	}

	body, err := generateBody(doc)
	if err != nil {
		return "", err
	}

	return formatIntoGoFile(pkgName, imports, body)
}

func GenerateStubCodesIntoFile(doc *ControllerManagerDocument, path string) error {
	fileName := doc.FileName[:strings.LastIndex(doc.FileName, ".")] + ".go"
	s, err := GenerateStubCodes(doc, filepath.Base(path))
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, fileName), []byte(s), 0644)
}
