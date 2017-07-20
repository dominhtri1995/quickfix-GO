# quickfix-GO

- This is QuickFix Code for TT FIX Adapter in Go. Documentation can be found here:

https://www.tradingtechnologies.com/help/fix-adapter-reference/introduction/whats-new-in-fix-adapter-717x/
https://godoc.org/github.com/quickfixgo/quickfix

- To get started :
```
$ go get github.com/quickfixgo/quickfix
$ git clone https://github.com/dominhtri1995/quickfix-GO $GOPATH/src/github.com.quickfixgo/examples/tradeclient
$ cd $GOPATH/src/github.com.quickfixgo/examples/tradeclient
$ go build
$ ./tradeclient
```

Wait for all the session get created and sequence number is reset. Read main.go files for examples of functionalities available through FIX. 

Usage:
```
ordStatus := TT_NewOrderSingle(xid.New().String(), "venustech", "1", "2", "500", "4570", "BZ", "CME", "201709", "FUT","VENUSTECH3")
			if(ordStatus.status == "ok"){
				fmt.Printf("Order Placed Successfully \n")
			}else if(ordStatus.status == "rejected"){
				fmt.Printf("Order Rejected \n")
				if(ordStatus.reason != ""){
					fmt.Printf("Reason: %s",ordStatus.reason)
				}
			}
```

Good luck!

