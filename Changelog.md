# Changelog


## 1.1.0 (2018/11/13)

### Feature
- Add contex support , put tarscurrent in context,for getting client ip ,port and so on.
- Add optional parameter for put context in request pacakge
- Add filter for writing plugin of tars service
- Add zipkin opentracing plugin
- Add support for protocol buffers


### Fix and enhancement.

- Change request package sbuffer field from vector<unsigned byte> to vector<byte>
- Fix stat report bug
- Getting Loglevel for remote configration
- Fix deadlock of getting routing infomation in extreme situation
- Improve goroutine pool 
- Fix occasionally panic problem because of the starting sequence of goroutines
- Golint most of the codes