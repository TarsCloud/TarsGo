#!/usr/bin/env python

# auth: abelqjli@tencent.com 2018/5/14
import os
import os.path
import argparse

taf2go = os.getcwd()+"/tars2go "
jce_home = "/home/tafjce/"
out_dir = "/tmp/fastjcetest/"


def gen_test_go_file_for_module(dirName,varList):
    code = "package " + dirName
    code +='''
    import (
    "testing"
    "tars/protocol/codec"
    "bytes"
    "reflect"
    "encoding/json"
    "math/rand"
    "strconv"
    "time"
    )
    
    type R_W_Func_I interface{
         ReadFrom(_is *codec.Reader) error
         WriteTo(_os *codec.Buffer) error
    }
    func Check(v R_W_Func_I,t *testing.T){
        w := codec.NewBuffer()
        v.WriteTo(w)
        v_bytes_1 := w.ToBytes()
        v_json_1 ,_:= json.Marshal(v)
        r := codec.NewReader(v_bytes_1)
        r_err := v.ReadFrom(r)
        w = codec.NewBuffer()
        v.WriteTo(w)
        v_bytes_2 := w.ToBytes()
        v_json_2 ,_:= json.Marshal(v)
        
        if !bytes.Equal(v_bytes_1,v_bytes_2) && !bytes.Equal(v_json_1,v_json_2) && len(v_bytes_1)!=len(v_bytes_2){
            t.Errorf("%s:en/decode error,readError:%v,len1:%d,len2:%d,json1:%s,json2:%s",
                     reflect.TypeOf(v).String(),r_err,len(v_bytes_1),len(v_bytes_2),string(v_json_1),string(v_json_2))
        }
    }
    func RandFillJceData(v reflect.Value){
        defer func(){
            if r := recover(); r != nil{
                //do nothing
                return
            }
        }()
        _RandFillJceData(v)
    }
    func _RandFillJceData(v reflect.Value){
        switch v.Kind(){
        case reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64,reflect.Int:
            v.SetInt(int64(rand.Int()))
        case reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64,reflect.Uint:
            v.SetUint(uint64(rand.Uint32()))
        case reflect.Float32:
            v.SetFloat(float64(rand.Float32()))
        case reflect.Float64:
            v.SetFloat(rand.Float64())
        case reflect.String:
            v.SetString(strconv.Itoa(rand.Int()))
        case reflect.Slice:
            v.Set(reflect.MakeSlice(v.Type().Elem(),10,10))
            num := rand.Intn(10)
            for num > 0{
                item := reflect.New(v.Type().Elem()).Elem()
                _RandFillJceData(item)
                v.Set(reflect.Append(v, item))
                num = num -1
            }
        case reflect.Map:
            v.Set(reflect.MakeMap(v.Type()))
            num := rand.Intn(10)
            for num > 0 {
                key := reflect.New(v.Type().Key()).Elem()
                _RandFillJceData(key)
                value := reflect.New(v.Type().Elem()).Elem()
                _RandFillJceData(value)
                v.SetMapIndex(key, value)
                num = num -1
            }
        case reflect.Struct:
            i := 0
            for i < v.NumField() {
                _RandFillJceData(v.Field(i))
                i = i + 1
        }
        }
    }
    func TestJce(t *testing.T){
        rand.Seed(time.Now().UnixNano())
    '''
    tpl = '''
        {TypeName}_Var := new({TypeName})
        RandFillJceData(reflect.ValueOf({TypeName}_Var).Elem())
        var {TypeName}_Var_I interface{}= {TypeName}_Var
        if val,ok := {TypeName}_Var_I.(interface{
             ReadFrom(_is *codec.Reader) error
             WriteTo(_os *codec.Buffer) error
        }); ok {
            Check(val,t)
        }
        '''
    for var in varList:
        code += tpl.replace("{TypeName}",var)
    code += "\n}"
    return code

def gen_test_go_file_for_all_module(dir_module):
    for parent,child_dirs,child_files in os.walk(dir_module):
        go_type_files = filter(lambda x:x.endswith(".go") and 
                                        not x.endswith("_IF.go") and 
                                        not x.endswith("_test.go") and 
                                        not x.endswith("_const.go"),child_files)
        go_types = list(map(lambda x:x.replace(".go",""),go_type_files))
        needGen = True
        if len(go_types) == 0:
            needGen = False
        testFName = os.path.join(parent,os.path.basename(parent) + "_Jce_test.go")     
        if needGen:
            with open(testFName,'w') as testFOut:
                testFOut.write(gen_test_go_file_for_module(os.path.basename(parent),go_types))

def gen_go_file_for_all_jce(dir_jce,dir_out):
    for parent,child_dirs,child_files in os.walk(dir_jce):
        os.system("cd "+parent +"&& " + taf2go + " -outdir=" + dir_out+ " *.jce")    
        
def excute_all_test_go_file(dir):
    for parent,child_dirs,child_files in os.walk(dir):
        os.system("cd "+parent +"&& go test")
        
if __name__ == "__main__":
    arg_parser = argparse.ArgumentParser(prog = 'valid', 
                                         description = 'test jce encode/decode')
    arg_parser.add_argument('-all', '--all' , action = 'store_true', 
                            default = False, help = 'do all action. 1.gen code for jce,2.gen test code ,3.and excute testing code')
    arg_parser.add_argument('-jce', '--jce' , action = 'store_true', 
                            default = False, help = 'generate golang code for jce')
    arg_parser.add_argument('-gentest', '--gentest' , action = 'store_true', 
                            default = False, help = 'generate test golang code for jce')
    arg_parser.add_argument('-test', '--test' , action = 'store_true', 
                            default = False, help = 'excute the golang testing code')                            
    arg_val_map = arg_parser.parse_args()
    if arg_val_map.all or arg_val_map.jce:
        gen_go_file_for_all_jce(jce_home,out_dir)
    if arg_val_map.all or arg_val_map.gentest:
        gen_test_go_file_for_all_module(out_dir)
    if arg_val_map.all or arg_val_map.test:
        excute_all_test_go_file(out_dir)
