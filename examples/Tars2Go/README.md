# 支持include参数 和 单个tars协议文件定义多个module
```bash
cd ../../tars/tools/tars2go && go install && cd -
tars2go -include=base -outdir=tars-protocol -module demo demo.tars
# 或者
go run ../../tars/tools/tars2go/*.go -include=base -outdir=tars-protocol -module demo demo.tars
```