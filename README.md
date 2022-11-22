
# GoBoom

A simple tool to "DDoS" a webserver via multiple threads and eachs using differents proxy



## Features

- Auto harvest of multiple sources for free proxy
- Possibility to add file with your own proxy (ip:port format)
- Possibility to set your own threads limit


## Usage/Examples
### Help
```shell
go run GoBoom.go -h

usage: GoBoom [-h|--help] -d|--domain "<value>" [-t|--threads "<value>"]
              [-p|--proxy-file "<value>" [-p|--proxy-file "<value>" ...]]

              Boom some website by proxy

Arguments:

  -h  --help        Print help information
  -d  --domain      Domain to boom
  -t  --threads     Number of threads. Default: max
  -p  --proxy-file  Proxy file(s), separate with a ',' each files. Format of
                    file(s) must be ip:port. Default: []
```
### With Golang
```shell
git clone https://github.com/ugomeguerditchian/GoBoom
cd GoBoom
go run GoBoom.go -d example.com
```
### With binaries
Open it in a terminal and add your args
```shell
GoBoom.exe -d example.com 

```
### Add your own proxy file
You can add your own file containing proxy :
```shell
    go run GoBoom.go -d example.com -p C:\myfile1.txt,C:\myfile2.txt

```



## Authors

- [@ugomeguerditchian](https://github.com/ugomeguerditchian)

## Contributors

- [@lisandro-git](https://github.com/lisandro-git)

