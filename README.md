# Landis
a go LAN discovery tool kits package

## contributor
MurInJ

## function
Get the address that are using this service in the same LAN

## install
```shell
go get github.com/MurInJ/Landis
```

## quick start
```go
port := 5003
s := Landis.NewLanDiscovery(port)
s.Start()
defer s.Close()
port2 := 5004
s2 := Landis.NewLanDiscovery(port2)
s2.Start()
defer s2.Close()
port3 := 5005
s3 := Landis.NewLanDiscovery(port3)
s3.Start()
defer s3.Close()

fmt.Println(s.List())
```