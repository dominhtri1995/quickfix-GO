package main

import (
	"fmt"
	"os"
	"bufio"
	"github.com/rs/xid"
	"os/exec"
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
			uan := TT_PAndLSOD(xid.New().String(), "venustech", "TTORDFA224222", "VENUSTECH3")
			if uan.status == "rejected" {
				fmt.Printf("Error when getting UAN: %s", uan.reason)
				continue
			}
			fmt.Printf("We have %d position in total for account %s and accountgroup %s \n", len(uan.reports), uan.account, uan.accountGroup)
			for _, uap := range uan.reports {
				fmt.Printf("%s \n", uap.product)
			}
		case "2":
			ordStatus := TT_NewOrderSingle(xid.New().String(), "venustech", "1", "2", "500", "4570", "BZ", "CME", "201709", "FUT", "1", "VENUSTECH3")
			if ordStatus.status == "ok" {
				fmt.Printf("Order %s %s at %s Placed Successfully \n", ordStatus.side, ordStatus.symbol, ordStatus.price)
			} else if ordStatus.status == "rejected" {
				fmt.Printf("Order Rejected \n")
				if ordStatus.reason != "" {
					fmt.Printf("Reason: %s", ordStatus.reason)
				}
			}
		case "3":
			wo := TT_WorkingOrder("venustech", "VENUSTECH3")
			if wo.status != "rejected" {
				for _, order := range wo.workingOrders {
					fmt.Printf("%s \n", order.symbol)
				}
			} else {
				fmt.Printf("Error when getting Working Orders: %s", wo.reason)
			}

		case "4":
			mdr := TT_MarketDataRequest(xid.New().String(), "0", 0, "2", "BZ", "CME", "201709", "FUT", "VENUSTECH3")
			fmt.Printf("Price :%s \n", mdr.price)

		case "5":
			wo := TT_WorkingOrder("venustech", "VENUSTECH3") // get all working order

			ordStatus := TT_OrderCancel(xid.New().String(), wo.workingOrders[0].orderID, "VENUSTECH3") // Cancel the first working order
			if ordStatus.status == "ok" {
				fmt.Printf("Order  %s %s at %s Cancelled Successfully \n", ordStatus.side, ordStatus.symbol, ordStatus.price)
			} else if ordStatus.status == "rejected" {
				fmt.Printf("Order Cancelled Rejected \n")
				if ordStatus.reason != "" {
					fmt.Printf("Reason: %s", ordStatus.reason)
				}
			}
		case "6":
			wo := TT_WorkingOrder("venustech", "VENUSTECH3") // get all working order
			for i := range wo.workingOrders {
				ordStatus := TT_OrderCancel(xid.New().String(), wo.workingOrders[i].orderID, "VENUSTECH3") // Cancel the first working order
				fmt.Printf("Status: %s \n", ordStatus.status)
			}
		case "7":
			wo := TT_WorkingOrder("venustech", "VENUSTECH3") // get all working order
			// Replace/Edit the first working order
			order := wo.workingOrders[0]
			//Change quantity of that working order to 962
			ordStatus := TT_OrderCancelReplace(order.orderID, xid.New().String(), wo.account, order.sideNum, order.ordType, "962", order.price, order.symbol, order.exchange, order.productMaturity, order.productType, order.timeInForce, "VENUSTECH3")
			if ordStatus.status == "ok" {
				fmt.Printf("Order Replaced Successfully: %s %s at %s \n", ordStatus.side, ordStatus.symbol, ordStatus.price)
			} else if ordStatus.status == "rejected" {
				fmt.Printf("Order Replaced Rejected \n")
				if ordStatus.reason != "" {
					fmt.Printf("Reason: %s", ordStatus.reason)
				}
			}
		case "8":
			//pass in parameters for filter, pass "" if do not want to use that criteria  //00A0IR00BZZ
			sdr := TT_QuerySecurityDefinitionRequest(xid.New().String(),"BZ","","","","VENUSTECH3")
			if sdr.status =="ok"{
				for _,security := range sdr.securityList{
					fmt.Printf("%s with tickValue: %f and tickSize %f %s %s\n",security.symbol,security.tickValue,security.tickSize,security.securityAltID,security.productMaturity)
				}
			}else{
				fmt.Println("error getting security definition")
				fmt.Println("Reason: ",sdr.reason)
			}
		case "9":
			break Loop
		default:
			cmd := exec.Command("clear") //Linux example, its tested
			cmd.Stdout = os.Stdout
			cmd.Run()
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
	fmt.Println("8) Test Security Definition Request")
	fmt.Println("9) Quit")
	fmt.Print("Action: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}
