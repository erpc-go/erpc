package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// 全局 map 避免重复生成
var (
	fileMap = make(map[string]bool, 0)
)

// GenGo record go code information.
type GenGo struct {
	I         []string // imports with path
	code      bytes.Buffer
	vc        int // var count. Used to generate unique variable names
	path      string
	codecPath string
	module    string
	prefix    string
	p         *Parse
}

// NewGenGo build up a new path
func NewGenGo(path string, module string, outdir string) *GenGo {
	if outdir != "" {
		b := []byte(outdir)
		last := b[len(b)-1:]
		if string(last) != "/" {
			outdir += "/"
		}
	}

	return &GenGo{
		I:    []string{},
		code: bytes.Buffer{},
		vc:   0,
		path: path,

		// 生成后的代码依赖的基础 codec 代码
		codecPath: "github.com/edte/erpc/codec/jce",
		module:    module,
		prefix:    outdir,
		p:         &Parse{},
	}
}

func path2ProtoName(path string) string {
	iBegin := strings.LastIndex(path, "/")
	if iBegin == -1 || iBegin >= len(path)-1 {
		iBegin = 0
	} else {
		iBegin++
	}
	iEnd := strings.LastIndex(path, ".jce")
	if iEnd == -1 {
		iEnd = len(path)
	}

	return path[iBegin:iEnd]
}

// Initial capitalization
func upperFirstLetter(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return strings.ToUpper(string(s[0]))
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func errString(hasRet bool) string {
	var retStr string
	if hasRet {
		retStr = "return ret, err"
	} else {
		retStr = "return err"
	}
	return `if err != nil {
  ` + retStr + `
  }` + "\n"
}

func genForHead(vc string) string {
	i := `i` + vc
	e := `e` + vc
	return ` for ` + i + `,` + e + ` := int32(0), length;` + i + `<` + e + `;` + i + `++ `
}

// === rename area ===
// 0. rename module
func (p *Parse) rename() {
	p.OriginModule = p.Module
	if moduleUpper {
		p.Module = upperFirstLetter(p.Module)
	}
}

// 1. struct rename
// struct Name { 1 require Mb type}
func (st *StructInfo) rename() {
	st.OriginName = st.Name
	st.Name = upperFirstLetter(st.Name)
	for i := range st.Mb {
		st.Mb[i].OriginKey = st.Mb[i].Key
		st.Mb[i].Key = upperFirstLetter(st.Mb[i].Key)
	}
}

func (en *EnumInfo) rename() {
	en.OriginName = en.Name
	en.Name = upperFirstLetter(en.Name)
	for i := range en.Mb {
		en.Mb[i].Key = upperFirstLetter(en.Mb[i].Key)
	}
}

func (cst *ConstInfo) rename() {
	cst.OriginName = cst.Name
	cst.Name = upperFirstLetter(cst.Name)
}

// 3. genType rename all Type

// === rename end ===

// Gen to parse file.
func (gen *GenGo) Gen() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			// set exit code
			os.Exit(1)
		}
	}()

	gen.p = ParseFile(gen.path, make([]string, 0))
	gen.genAll()
}

func (gen *GenGo) genAll() {
	if fileMap[gen.path] {
		// already compiled
		return
	}

	gen.p.rename()
	gen.genInclude(gen.p.IncParse)

	gen.code.Reset()
	gen.genHead()
	gen.genPackage()

	for _, v := range gen.p.Enum {
		gen.genEnum(&v)
	}

	gen.genConst(gen.p.Const)

	for _, v := range gen.p.Struct {
		gen.genStruct(&v)
	}

	if len(gen.p.Enum) > 0 || len(gen.p.Const) > 0 || len(gen.p.Struct) > 0 {
		gen.saveToSourceFile(path2ProtoName(gen.path) + ".jce.go")
	}

	fileMap[gen.path] = true
}

func (gen *GenGo) genErr(err string) {
	panic(err)
}

func (gen *GenGo) saveToSourceFile(filename string) {
	var beauty []byte
	var err error
	prefix := gen.prefix

	beauty, err = format.Source(gen.code.Bytes())
	if err != nil {
		if debug {
			fmt.Println("------------------")
			fmt.Println(string(gen.code.Bytes()))
			fmt.Println("------------------")
		}
		gen.genErr("go fmt fail. " + filename + " " + err.Error())
	}

	if filename == "stdout" {
		fmt.Println(string(beauty))
	} else {
		var mkPath string
		if !moduleCycle {
			mkPath = prefix + gen.p.Module
		}
		err = os.MkdirAll(mkPath, 0766)

		if err != nil {
			gen.genErr(err.Error())
		}
		err = ioutil.WriteFile(mkPath+"/"+filename, beauty, 0666)

		if err != nil {
			gen.genErr(err.Error())
		}
	}
}

func (gen *GenGo) genVariableName(prefix, name string) string {
	if prefix != "" {
		return prefix + name
	} else {
		return strings.Trim(name, "()")
	}
}

func (gen *GenGo) genHead() {
	gen.code.WriteString(`// Code generated by jce2go` + VERSION + `. DO NOT EDIT.
// source: ` + filepath.Base(gen.path) + `
`)
}

func (gen *GenGo) genPackage() {
	gen.code.WriteString("package " + gen.p.Module + "\n\n")
	gen.code.WriteString(`
import (
	"fmt"
    "io"

`)
	gen.code.WriteString("\"" + gen.codecPath + "\"\n")

	mImports := make(map[string]bool)
	for _, st := range gen.p.Struct {
		if moduleCycle == true {
			for k, v := range st.DependModuleWithJce {
				gen.genStructImport(k, v, mImports)
			}
		} else {
			for k := range st.DependModule {
				gen.genStructImport(k, "", mImports)
			}
		}
	}
	for path := range mImports {
		gen.code.WriteString(path + "\n")
	}

	gen.code.WriteString(`)

	// Reference imports to suppress errors if they are not otherwise used.
	var _ = fmt.Errorf
	var _ = jce.FromInt8

`)
}

func (gen *GenGo) genStructImport(module string, protoName string, mImports map[string]bool) {
	var moduleStr string
	var jcePath string
	var moduleAlia string
	if moduleCycle == true {
		moduleStr = module[len(protoName)+1:]
		jcePath = protoName + "/"
		moduleAlia = module + " "
	} else {
		moduleStr = module
	}

	for _, p := range gen.I {
		if strings.HasSuffix(p, "/"+moduleStr) {
			mImports[`"`+p+`"`] = true
			return
		}
	}

	if moduleUpper {
		moduleAlia = upperFirstLetter(moduleAlia)
	}

	// example:
	// TarsTest.tars, MyApp
	// gomod:
	// github.com/xxx/yyy/tars-protocol/MyApp
	// github.com/xxx/yyy/tars-protocol/TarsTest/MyApp
	//
	// gopath:
	// MyApp
	// TarsTest/MyApp
	var modulePath string
	if gen.module != "" {
		mf := filepath.Clean(filepath.Join(gen.module, gen.prefix))
		if runtime.GOOS == "windows" {
			mf = strings.ReplaceAll(mf, string(os.PathSeparator), string('/'))
		}
		modulePath = fmt.Sprintf("%s/%s%s", mf, jcePath, moduleStr)
	} else {
		modulePath = fmt.Sprintf("%s%s", jcePath, moduleStr)
	}
	mImports[moduleAlia+`"`+modulePath+`"`] = true
}

func (gen *GenGo) genIFImport(module string, protoName string) {
	var moduleStr string
	var jcePath string
	var moduleAlia string
	if moduleCycle == true {
		moduleStr = module[len(protoName)+1:]
		jcePath = protoName + "/"
		moduleAlia = module + " "
	} else {
		moduleStr = module
	}
	for _, p := range gen.I {
		if strings.HasSuffix(p, "/"+moduleStr) {
			gen.code.WriteString(`"` + p + `"` + "\n")
			return
		}
	}

	if moduleUpper {
		moduleAlia = upperFirstLetter(moduleAlia)
	}

	// example:
	// TarsTest.tars, MyApp
	// gomod:
	// github.com/xxx/yyy/tars-protocol/MyApp
	// github.com/xxx/yyy/tars-protocol/TarsTest/MyApp
	//
	// gopath:
	// MyApp
	// TarsTest/MyApp
	var modulePath string
	if gen.module != "" {
		mf := filepath.Clean(filepath.Join(gen.module, gen.prefix))
		if runtime.GOOS == "windows" {
			mf = strings.ReplaceAll(mf, string(os.PathSeparator), string('/'))
		}
		modulePath = fmt.Sprintf("%s/%s%s", mf, jcePath, moduleStr)
	} else {
		modulePath = fmt.Sprintf("%s%s", jcePath, moduleStr)
	}
	gen.code.WriteString(moduleAlia + `"` + modulePath + `"` + "\n")
}

func (gen *GenGo) genType(ty *VarType) string {
	ret := ""
	switch ty.Type {
	case tkTBool:
		ret = "bool"
	case tkTInt:
		if ty.Unsigned {
			ret = "uint32"
		} else {
			ret = "int32"
		}
	case tkTShort:
		if ty.Unsigned {
			ret = "uint16"
		} else {
			ret = "int16"
		}
	case tkTByte:
		if ty.Unsigned {
			ret = "uint8"
		} else {
			ret = "int8"
		}
	case tkTLong:
		if ty.Unsigned {
			ret = "uint64"
		} else {
			ret = "int64"
		}
	case tkTFloat:
		ret = "float32"
	case tkTDouble:
		ret = "float64"
	case tkTString:
		ret = "string"
	case tkTVector:
		ret = "[]" + gen.genType(ty.TypeK)
	case tkTMap:
		ret = "map[" + gen.genType(ty.TypeK) + "]" + gen.genType(ty.TypeV)
	case tkName:
		ret = strings.Replace(ty.TypeSt, "::", ".", -1)
		vec := strings.Split(ty.TypeSt, "::")
		for i := range vec {
			if moduleUpper {
				vec[i] = upperFirstLetter(vec[i])
			} else {
				if i == (len(vec) - 1) {
					vec[i] = upperFirstLetter(vec[i])
				}
			}
		}
		ret = strings.Join(vec, ".")
	case tkTArray:
		ret = "[" + fmt.Sprintf("%v", ty.TypeL) + "]" + gen.genType(ty.TypeK)
	default:
		gen.genErr("Unknown Type " + TokenMap[ty.Type])
	}
	return ret
}

func (gen *GenGo) genStructDefine(st *StructInfo) {
	c := &gen.code
	c.WriteString("// " + st.Name + " struct implement\n")
	c.WriteString("type " + st.Name + " struct {\n")

	for _, v := range st.Mb {
		if jsonOmitEmpty {
			c.WriteString("\t" + v.Key + " " + gen.genType(v.Type) + " `json:\"" + v.OriginKey + ",omitempty\"`\n")
		} else {
			c.WriteString("\t" + v.Key + " " + gen.genType(v.Type) + " `json:\"" + v.OriginKey + "\"`\n")
		}
	}
	c.WriteString("}\n")
}

func (gen *GenGo) genFunResetDefault(st *StructInfo) {
	c := &gen.code

	c.WriteString("func (st *" + st.Name + ") ResetDefault() {\n")

	for _, v := range st.Mb {
		if v.Type.CType == tkStruct {
			c.WriteString("st." + v.Key + ".ResetDefault()\n")
		}
		if v.Default == "" {
			continue
		}
		c.WriteString("st." + v.Key + " = " + v.Default + "\n")
	}
	c.WriteString("}\n")
}

func (gen *GenGo) genWriteSimpleList(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	tag := strconv.Itoa(int(mb.Tag))
	unsign := "Int8"
	if mb.Type.TypeK.Unsigned {
		unsign = "Uint8"
	}
	errStr := errString(hasRet)
	c.WriteString(`
err = buf.WriteHead(jce.SimpleList, ` + tag + `)
` + errStr + `
err = buf.WriteHead(jce.BYTE, 0)
` + errStr + `
err = buf.WriteInt32(int32(len(` + gen.genVariableName(prefix, mb.Key) + `)), 0)
` + errStr + `
err = buf.WriteSlice` + unsign + `(` + gen.genVariableName(prefix, mb.Key) + `)
` + errStr + `
`)
}

func (gen *GenGo) genWriteVector(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code

	// SimpleList
	if mb.Type.TypeK.Type == tkTByte && !mb.Type.TypeK.Unsigned {
		gen.genWriteSimpleList(mb, prefix, hasRet)
		return
	}
	errStr := errString(hasRet)

	// LIST
	tag := strconv.Itoa(int(mb.Tag))
	c.WriteString(`
err = buf.WriteHead(jce.LIST, ` + tag + `)
` + errStr + `
err = buf.WriteInt32(int32(len(` + gen.genVariableName(prefix, mb.Key) + `)), 0)
` + errStr + `
for _, v := range ` + gen.genVariableName(prefix, mb.Key) + ` {
`)
	// for _, v := range can nesting for _, v := range，does not conflict, support multidimensional arrays

	dummy := &StructMember{}
	dummy.Type = mb.Type.TypeK
	dummy.Key = "v"
	gen.genWriteVar(dummy, "", hasRet)

	c.WriteString("}\n")
}

func (gen *GenGo) genWriteArray(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code

	// SimpleList
	if mb.Type.TypeK.Type == tkTByte && !mb.Type.TypeK.Unsigned {
		gen.genWriteSimpleList(mb, prefix, hasRet)
		return
	}
	errStr := errString(hasRet)

	// LIST
	tag := strconv.Itoa(int(mb.Tag))
	c.WriteString(`
err = buf.WriteHead(jce.LIST, ` + tag + `)
` + errStr + `
err = buf.WriteInt32(int32(len(` + gen.genVariableName(prefix, mb.Key) + `)), 0)
` + errStr + `
for _, v := range ` + gen.genVariableName(prefix, mb.Key) + ` {
`)
	// for _, v := range can nesting for _, v := range，does not conflict, support multidimensional arrays

	dummy := &StructMember{}
	dummy.Type = mb.Type.TypeK
	dummy.Key = "v"
	gen.genWriteVar(dummy, "", hasRet)

	c.WriteString("}\n")
}

func (gen *GenGo) genWriteStruct(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	tag := strconv.Itoa(int(mb.Tag))
	c.WriteString(`
err = ` + prefix + mb.Key + `.WriteBlock(buf, ` + tag + `)
` + errString(hasRet) + `
`)
}

func (gen *GenGo) genWriteMap(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	tag := strconv.Itoa(int(mb.Tag))
	vc := strconv.Itoa(gen.vc)
	gen.vc++
	errStr := errString(hasRet)
	c.WriteString(`
err = buf.WriteHead(jce.MAP, ` + tag + `)
` + errStr + `
err = buf.WriteInt32(int32(len(` + gen.genVariableName(prefix, mb.Key) + `)), 0)
` + errStr + `
for k` + vc + `, v` + vc + ` := range ` + gen.genVariableName(prefix, mb.Key) + ` {
`)
	// for _, v := range can nesting for _, v := range，does not conflict, support multidimensional arrays

	dummy := &StructMember{}
	dummy.Type = mb.Type.TypeK
	dummy.Key = "k" + vc
	gen.genWriteVar(dummy, "", hasRet)

	dummy = &StructMember{}
	dummy.Type = mb.Type.TypeV
	dummy.Key = "v" + vc
	dummy.Tag = 1
	gen.genWriteVar(dummy, "", hasRet)

	c.WriteString("}\n")
}

func (gen *GenGo) genWriteVar(v *StructMember, prefix string, hasRet bool) {
	c := &gen.code

	switch v.Type.Type {
	case tkTVector:
		gen.genWriteVector(v, prefix, hasRet)
	case tkTArray:
		gen.genWriteArray(v, prefix, hasRet)
	case tkTMap:
		gen.genWriteMap(v, prefix, hasRet)
	case tkName:
		if v.Type.CType == tkEnum {
			// tkEnum enumeration processing
			tag := strconv.Itoa(int(v.Tag))
			c.WriteString(`
err = buf.WriteInt32(int32(` + gen.genVariableName(prefix, v.Key) + `),` + tag + `)
` + errString(hasRet) + `
`)
		} else {
			gen.genWriteStruct(v, prefix, hasRet)
		}
	default:
		tag := strconv.Itoa(int(v.Tag))
		c.WriteString(`
err = buf.Write` + upperFirstLetter(gen.genType(v.Type)) + `(` + gen.genVariableName(prefix, v.Key) + `, ` + tag + `)
` + errString(hasRet) + `
`)
	}
}

func (gen *GenGo) genFunWriteBlock(st *StructInfo) {
	c := &gen.code

	// WriteBlock function head
	c.WriteString(`// WriteBlock encode struct
func (st *` + st.Name + `) WriteBlock(w io.Writer, tag byte) (err error) {
    buf := jce.NewEncoder(w)

	if err = buf.WriteHead(jce.StructBegin, tag);err != nil {
        return
    }

	if err = st.WriteTo(w);err != nil {
        return
    }

	if err = buf.WriteHead(jce.StructEnd, 0);err!=nil {
        return 
    }

	return
}
`)
}

func (gen *GenGo) genFunWriteTo(st *StructInfo) {
	c := &gen.code

	c.WriteString(`// WriteTo encode struct to buffer
func (st *` + st.Name + `) WriteTo(w io.Writer) (err error) {
    buf := jce.NewEncoder(w)
`)
	for _, v := range st.Mb {
		gen.genWriteVar(&v, "st.", false)
	}

	c.WriteString(`
    buf.Flush()
	return err
}
`)
}

func (gen *GenGo) genReadSimpleList(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	unsign := "Int8"
	if mb.Type.TypeK.Unsigned {
		unsign = "Uint8"
	}
	errStr := errString(hasRet)

	c.WriteString(`
_, err = readBuf.SkipTo(jce.BYTE, 0, true)
` + errStr + `
err = readBuf.ReadInt32(&length, 0, true)
` + errStr + `
err = readBuf.ReadSlice` + unsign + `(&` + prefix + mb.Key + `, length, true)
` + errStr + `
`)
}

func (gen *GenGo) genReadVector(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	errStr := errString(hasRet)

	// LIST
	tag := strconv.Itoa(int(mb.Tag))
	vc := strconv.Itoa(gen.vc)
	gen.vc++
	require := "false"
	if mb.Require {
		require = "true"
	}
	if require == "false" {
		c.WriteString(`
have, ty, err = readBuf.SkipToNoCheck(` + tag + `,` + require + `)
` + errStr + `
if have {`)
	} else {
		c.WriteString(`
_, ty, err = readBuf.SkipToNoCheck(` + tag + `,` + require + `)
` + errStr + `
`)
	}

	c.WriteString(`
if ty == jce.LIST {
	err = readBuf.ReadInt32(&length, 0, true)
  ` + errStr + `
  ` + gen.genVariableName(prefix, mb.Key) + ` = make(` + gen.genType(mb.Type) + `, length)
  ` + genForHead(vc) + `{
`)

	dummy := &StructMember{}
	dummy.Type = mb.Type.TypeK
	dummy.Key = mb.Key + "[i" + vc + "]"
	gen.genReadVar(dummy, prefix, hasRet)

	c.WriteString(`}
} else if ty == jce.SimpleList {
`)
	if mb.Type.TypeK.Type == tkTByte {
		gen.genReadSimpleList(mb, prefix, hasRet)
	} else {
		c.WriteString(`err = fmt.Errorf("not support SimpleList type")
    ` + errStr)
	}
	c.WriteString(`
} else {
  err = fmt.Errorf("require vector, but not")
  ` + errStr + `
}
`)

	if require == "false" {
		c.WriteString("}\n")
	}
}

func (gen *GenGo) genReadArray(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	errStr := errString(hasRet)

	// LIST
	tag := strconv.Itoa(int(mb.Tag))
	vc := strconv.Itoa(gen.vc)
	gen.vc++
	require := "false"
	if mb.Require {
		require = "true"
	}

	if require == "false" {
		c.WriteString(`
have, ty, err = readBuf.SkipToNoCheck(` + tag + `,` + require + `)
` + errStr + `
if have {`)
	} else {
		c.WriteString(`
_, ty, err = readBuf.SkipToNoCheck(` + tag + `,` + require + `)
` + errStr + `
`)
	}

	c.WriteString(`
if ty == jce.LIST {
	err = readBuf.ReadInt32(&length, 0, true)
  ` + errStr + `
  ` + genForHead(vc) + `{
`)

	dummy := &StructMember{}
	dummy.Type = mb.Type.TypeK
	dummy.Key = mb.Key + "[i" + vc + "]"
	gen.genReadVar(dummy, prefix, hasRet)

	c.WriteString(`}
} else if ty == jce.SimpleList {
`)
	if mb.Type.TypeK.Type == tkTByte {
		gen.genReadSimpleList(mb, prefix, hasRet)
	} else {
		c.WriteString(`err = fmt.Errorf("not support SimpleList type")
    ` + errStr)
	}
	c.WriteString(`
} else {
  err = fmt.Errorf("require array, but not")
  ` + errStr + `
}
`)

	if require == "false" {
		c.WriteString("}\n")
	}
}

func (gen *GenGo) genReadStruct(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	tag := strconv.Itoa(int(mb.Tag))
	require := "false"
	if mb.Require {
		require = "true"
	}
	c.WriteString(`
err = ` + prefix + mb.Key + `.ReadBlock(readBuf, ` + tag + `, ` + require + `)
` + errString(hasRet) + `
`)
}

func (gen *GenGo) genReadMap(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	tag := strconv.Itoa(int(mb.Tag))
	errStr := errString(hasRet)
	vc := strconv.Itoa(gen.vc)
	gen.vc++
	require := "false"
	if mb.Require {
		require = "true"
	}

	if require == "false" {
		c.WriteString(`
have, err = readBuf.SkipTo(jce.MAP, ` + tag + `, ` + require + `)
` + errStr + `
if have {`)
	} else {
		c.WriteString(`
_, err = readBuf.SkipTo(jce.MAP, ` + tag + `, ` + require + `)
` + errStr + `
`)
	}

	c.WriteString(`
err = readBuf.ReadInt32(&length, 0, true)
` + errStr + `
` + gen.genVariableName(prefix, mb.Key) + ` = make(` + gen.genType(mb.Type) + `)
` + genForHead(vc) + `{
	var k` + vc + ` ` + gen.genType(mb.Type.TypeK) + `
	var v` + vc + ` ` + gen.genType(mb.Type.TypeV) + `
`)

	dummy := &StructMember{}
	dummy.Type = mb.Type.TypeK
	dummy.Key = "k" + vc
	gen.genReadVar(dummy, "", hasRet)

	dummy = &StructMember{}
	dummy.Type = mb.Type.TypeV
	dummy.Key = "v" + vc
	dummy.Tag = 1
	gen.genReadVar(dummy, "", hasRet)

	c.WriteString(`
	` + prefix + mb.Key + `[k` + vc + `] = v` + vc + `
}
`)
	if require == "false" {
		c.WriteString("}\n")
	}
}

func (gen *GenGo) genReadVar(v *StructMember, prefix string, hasRet bool) {
	c := &gen.code

	switch v.Type.Type {
	case tkTVector:
		gen.genReadVector(v, prefix, hasRet)
	case tkTArray:
		gen.genReadArray(v, prefix, hasRet)
	case tkTMap:
		gen.genReadMap(v, prefix, hasRet)
	case tkName:
		if v.Type.CType == tkEnum {
			tag := strconv.Itoa(int(v.Tag))
			require := "false"
			if v.Require {
				require = "true"
			}
			c.WriteString(`
err = readBuf.ReadInt32((*int32)(&` + prefix + v.Key + `),` + tag + `, ` + require + `)
` + errString(hasRet) + `
`)
		} else {
			gen.genReadStruct(v, prefix, hasRet)
		}
	default:
		tag := strconv.Itoa(int(v.Tag))
		require := "false"
		if v.Require {
			require = "true"
		}
		c.WriteString(`
err = readBuf.Read` + upperFirstLetter(gen.genType(v.Type)) + `(&` + prefix + v.Key + `, ` + tag + `, ` + require + `)
` + errString(hasRet) + `
`)
	}
}

func (gen *GenGo) genFunReadFrom(st *StructInfo) {
	c := &gen.code

	c.WriteString(`// ReadFrom reads  from readBuf and put into struct.
func (st *` + st.Name + `) ReadFrom(r io.Reader) (err error) {
	var (
		length int32
		have bool
		ty byte
	)

    readBuf := jce.NewDecoder(r)
	st.ResetDefault()

`)

	for _, v := range st.Mb {
		gen.genReadVar(&v, "st.", false)
	}

	c.WriteString(`
	_ = err
	_ = length
	_ = have
	_ = ty
	return nil
}
`)
}

func (gen *GenGo) genFunReadBlock(st *StructInfo) {
	c := &gen.code

	c.WriteString(`// ReadBlock reads struct from the given tag , require or optional.
func (st *` + st.Name + `) ReadBlock(r io.Reader, tag byte, require bool) error {
	var (
		err error
		have bool
	)

    readBuf := jce.NewDecoder(r)

	st.ResetDefault()

	have, err = readBuf.SkipTo(jce.StructBegin, tag, require)
	if err != nil {
		return err
	}

	if !have {
		if require {
			return fmt.Errorf("require ` + st.Name + `, but not exist. tag %d", tag)
		}
		return nil
	}

  	err = st.ReadFrom(r)
  	if err != nil {
		return err
	}

	err = readBuf.SkipToStructEnd()
	if err != nil {
		return err
	}

	_ = have
	return nil
}
`)
}

func (gen *GenGo) genStruct(st *StructInfo) {
	gen.vc = 0
	st.rename()

	gen.genStructDefine(st)
	gen.genFunResetDefault(st)

	gen.genFunReadFrom(st)
	gen.genFunReadBlock(st)

	gen.genFunWriteTo(st)
	gen.genFunWriteBlock(st)
}

func (gen *GenGo) makeEnumName(en *EnumInfo, mb *EnumMember) string {
	return upperFirstLetter(en.Name) + "_" + upperFirstLetter(mb.Key)
}

func (gen *GenGo) genEnum(en *EnumInfo) {
	if len(en.Mb) == 0 {
		return
	}

	en.rename()

	c := &gen.code
	c.WriteString("type " + en.Name + " int32\n")
	c.WriteString("const (\n")
	var it int32
	for _, v := range en.Mb {
		if v.Type == 0 {
			//use value
			c.WriteString(gen.makeEnumName(en, &v) + ` = ` + strconv.Itoa(int(v.Value)) + "\n")
			it = v.Value + 1
		} else if v.Type == 1 {
			// use name
			find := false
			for _, ref := range en.Mb {
				if ref.Key == v.Name {
					find = true
					c.WriteString(gen.makeEnumName(en, &v) + ` = ` + gen.makeEnumName(en, &ref) + "\n")
					it = ref.Value + 1
					break
				}
				if ref.Key == v.Key {
					break
				}
			}
			if !find {
				gen.genErr(v.Name + " not define before use.")
			}
		} else {
			// use auto add
			c.WriteString(gen.makeEnumName(en, &v) + ` = ` + strconv.Itoa(int(it)) + "\n")
			it++
		}

	}

	c.WriteString(")\n")
}

func (gen *GenGo) genConst(cst []ConstInfo) {
	if len(cst) == 0 {
		return
	}

	c := &gen.code
	c.WriteString("//const as define in jce file\n")
	c.WriteString("const (\n")

	for _, v := range gen.p.Const {
		v.rename()
		c.WriteString(v.Name + " " + gen.genType(v.Type) + " = " + v.Value + "\n")
	}

	c.WriteString(")\n")
}

func (gen *GenGo) genInclude(ps []*Parse) {
	for _, v := range ps {
		gen2 := &GenGo{
			path:      v.Source,
			module:    gen.module,
			prefix:    gen.prefix,
			codecPath: gen.codecPath,
		}
		gen2.p = v
		gen2.genAll()
	}
}
