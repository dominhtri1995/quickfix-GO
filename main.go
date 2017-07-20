package main

import (
	"fmt"
	"os"
	"bufio"
	"github.com/rs/xid"
)

func main() {

	StartQuickFix()
Loop:
	for {
		action, err := QueryAction()
		if err != nil {
			break
		}
		switch action {
		case "1":
			uan := TT_PAndLSOD(xid.New().String(),"venustech" ,"TTORDFA222222","VENUSTECH3")
			if(uan.status == "rejected"){
				fmt.Printf("Error when getting UAN: %s",uan.reason)
				continue
			}
			fmt.Printf("We have %d position in total for account %s and accountgroup %s \n",len(uan.reports),uan.account,uan.accountGroup)
			for _, uap := range uan.reports {
				fmt.Printf("%s \n", uap.product)
			}
		case "2":
			ordStatus := TT_NewOrderSingle(xid.New().String(), "venustech", "1", "2", "500", "4570", "BZ", "CME", "201709", "FUT","VENUSTECH3")
			if(ordStatus.status == "ok"){
				fmt.Printf("Order Placed Successfully \n")
			}else if(ordStatus.status == "rejected"){
				fmt.Printf("Order Rejected \n")
				if(ordStatus.reason != ""){
					fmt.Printf("Reason: %s",ordStatus.reason)
				}
			}
		case "3":
			wo := TT_WorkingOrder("venustech","VENUSTECH3")
			if(wo.status != "rejected"){
				for _, order := range wo.workingOrders {
					fmt.Printf("%s \n", order.symbol)
				}
			}else {
				fmt.Printf("Error when getting Working Orders: %s",wo.reason)
			}

		case "4":
			mdr := TT_MarketDataRequest(xid.New().String(), "0", 0, "2", "BZ", "CME", "201709", "FUT","VENUSTECH")
			fmt.Printf("Price :%s \n", mdr.price)

		case "5":
			wo := TT_WorkingOrder("venustech","VENUSTECH3") // get all working order

			ordStatus := TT_OrderCancel(xid.New().String(), wo.workingOrders[0].orderID,"VENUSTECH3") // Cancel the first working order
			if(ordStatus.status == "ok"){
				fmt.Printf("Order Cancelled Successfully \n")
			}else if(ordStatus.status == "rejected"){
				fmt.Printf("Order Cancelled Rejected \n")
				if(ordStatus.reason != ""){
					fmt.Printf("Reason: %s",ordStatus.reason)
				}
			}
		case "6":
			wo := TT_WorkingOrder("venustech","VENUSTECH3") // get all working order
			for i := range wo.workingOrders{
				ordStatus := TT_OrderCancel(xid.New().String(), wo.workingOrders[i].orderID,"VENUSTECH3") // Cancel the first working order
				fmt.Printf("Status: %s \n", ordStatus.status)
			}
		case "7":
			wo := TT_WorkingOrder("venustech","VENUSTECH3") // get all working order
			// Replace/Edit the first working order
			order := wo.workingOrders[0]
			//Change quantity
			ordStatus := TT_OrderCancelReplace(order.orderID, xid.New().String(), wo.account, order.sideNum, order.ordType, "962", order.price, order.symbol, order.exchange, order.productMaturity, order.productType,"VENUSTECH3")
			if(ordStatus.status == "ok"){
				fmt.Printf("Order Replaced Successfully \n")
			}else if(ordStatus.status == "rejected"){
				fmt.Printf("Order Replaced Rejected \n")
				if(ordStatus.reason != ""){
					fmt.Printf("Reason: %s",ordStatus.reason)
				}
			}

		case "9":
			break Loop

		default:
			err = fmt.Errorf("unknown action: '%v'", action)
		}

		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}
}

func QueryAction() (string, error) {
	fmt.Println()
	fmt.Println("1) Get P&L")
	fmt.Println("2) Place new Order")
	fmt.Println("3) Get Working Order")
	fmt.Println("4) Market Data")
	fmt.Println("5) Cancel First working order")
	fmt.Println("6) Cancel All working order")
	fmt.Println("7) Replace First working order")
	fmt.Println("9) Quit")
	fmt.Print("Action: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}
