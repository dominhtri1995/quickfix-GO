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
			uan := TT_PAndLSOD(xid.New().String(), "venustech", "TTORDFA224222", "VENUSTECH")
			if uan.Status == "rejected" {
				fmt.Printf("Error when getting UAN: %s", uan.Reason)
				continue
			}
			fmt.Printf("We have %d position in total for Account %s and accountgroup %s \n", len(uan.Reports), uan.Account, uan.AccountGroup)
			for _, uap := range uan.Reports {
				fmt.Printf("%s %s %s\n",uap.Side, uap.SecurityAltID,uap.Price)
			}
		case "2":
			scanner := bufio.NewScanner(os.Stdin)

			fmt.Print("Side: ")
			scanner.Scan()
			side := scanner.Text()

			fmt.Print("Order type: ")
			scanner.Scan()
			ordType:= scanner.Text()

			fmt.Print("Quantity: ")
			scanner.Scan()
			qty:= scanner.Text()

			fmt.Print("Limit Price: ")
			scanner.Scan()
			limitprice := scanner.Text()

			fmt.Print("Stop Price: ")
			scanner.Scan()
			stopprice:= scanner.Text()

			fmt.Print("Symbol: ")
			scanner.Scan()
			symbol:= scanner.Text()

			fmt.Print("Exchange: ")
			scanner.Scan()
			exchange := scanner.Text()

			fmt.Print("Maturity: ")
			scanner.Scan()
			maturiy := scanner.Text()

			ordStatus := TT_NewOrderSingle(xid.New().String(), "venustech","tri", side, ordType, qty, limitprice,stopprice, symbol, exchange, maturiy, "FUT", "1","0","0", "B1","VENUSTECH3","TTDEV18O")
			if ordStatus.Status == "ok" {
				fmt.Println(ordStatus.Id)
				fmt.Printf("Order %s %s at %s Placed Successfully \n", ordStatus.Side, ordStatus.Symbol, ordStatus.Price)
			} else if ordStatus.Status == "rejected" {
				fmt.Printf("Order Rejected \n")
				if ordStatus.Reason != "" {
					fmt.Printf("Reason: %s", ordStatus.Reason)
				}
			}
		case "3":
			wo := TT_WorkingOrder("venustech2", "VENUSTECH9")
			if wo.Status != "rejected" {
				fmt.Printf("Number of order %d \n",len(wo.WorkingOrders))
				for _, order := range wo.WorkingOrders {
					fmt.Printf("%s \n", order.Symbol)
					fmt.Printf("limit p: %s\n",order.Price)
					fmt.Printf("qty: %s\n",order.Quantity)
					fmt.Printf ("filled %s\n",order.FilledQuantity)
					for _,u := range order.NoRelatedSymGroup{
						fmt.Printf("%s %s %s\n",u.UnderlyingSymbol,u.UnderlyingSecurityExchange, u.UnderlyingSecurityAltID)
					}
				}
			} else {
				fmt.Printf("Error when getting Working Orders: %s", wo.Reason)
			}

		case "4":
			mdr := TT_MarketDataRequest(xid.New().String(), "0", 0, "2", "BZ", "CME", "201709", "FUT", "VENUSTECH3")
			fmt.Printf("Price :%s \n", mdr.Price)

		case "5":
			wo := TT_WorkingOrder("venustech", "VENUSTECH3") // get all working order

			ordStatus := TT_OrderCancel(xid.New().String(), wo.WorkingOrders[0].OrderID, "VENUSTECH") // Cancel the first working order
			if ordStatus.Status == "ok" {
				fmt.Printf("Order  %s %s at %s Cancelled Successfully \n", ordStatus.Side, ordStatus.Symbol, ordStatus.Price)
			} else if ordStatus.Status == "rejected" {
				fmt.Printf("Order Cancelled Rejected \n")
				if ordStatus.Reason != "" {
					fmt.Printf("Reason: %s", ordStatus.Reason)
				}
			}
		case "6":
			wo := TT_WorkingOrder("venustech", "VENUSTECH3") // get all working order
			for i := range wo.WorkingOrders {
				ordStatus := TT_OrderCancel(xid.New().String(), wo.WorkingOrders[i].OrderID, "VENUSTECH3") // Cancel the first working order
				fmt.Printf("Status: %s \n", ordStatus.Status)
			}
		case "7":
			wo := TT_WorkingOrder("venustech", "VENUSTECH") // get all working order
			// Replace/Edit the first working order
			order := wo.WorkingOrders[0]
			//Change Quantity of that working order to 962
			ordStatus := TT_OrderCancelReplace(order.OrderID, xid.New().String(), wo.Account, order.SideNum, order.OrdType, "962", order.Price,order.StopPrice, order.Symbol, order.Exchange, order.ProductMaturity, order.ProductType, order.TimeInForce,order.StrikePrice,order.PutOrCall, "VENUSTECH")
			if ordStatus.Status == "ok" {
				fmt.Printf("Order Replaced Successfully: %s %s at %s \n", ordStatus.Side, ordStatus.Symbol, ordStatus.Price)
			} else if ordStatus.Status == "rejected" {
				fmt.Printf("Order Replaced Rejected \n")
				if ordStatus.Reason != "" {
					fmt.Printf("Reason: %s", ordStatus.Reason)
				}
			}
		case "8":
			//pass in parameters for filter, pass "" if do not want to use that criteria  //00A0IR00BZZ
			sdr := TT_QuerySecurityDefinitionRequest(xid.New().String(),"GE","CME","","OPT","VENUSTECH")
			if sdr.Status =="ok"{
				for _,security := range sdr.SecurityList {
					fmt.Printf("%s with TickValue: %f and TickSize %f %s %s %s\n",security.Symbol,security.TickValue,security.TickSize,security.SecurityAltID,security.ProductMaturity, security.Exchange)
				}
			}else{
				fmt.Println("error getting security definition")
				fmt.Printf("Reason: %s \n",sdr.Reason)
			}
		case "9":
			uan:= TT_Fills(xid.New().String(),"venustech3","VENUSTECH")
			if uan.Status == "ok"{
				fmt.Println("Get Fill ok")
				for _, uap := range uan.Reports {
					fmt.Printf("Fill: %s %s at %s quant: %s\n",uap.Side, uap.SecurityAltID,uap.Price,uap.Quantity)
				}
			}else{
				fmt.Println("get fill failed")
			}

		case "99":
			break Loop
		case "19"://test ordertype
			numbers := []string{"5","B","J","O","Q","R"}
			for _,n := range numbers{
				ordStatus := TT_NewOrderSingle(xid.New().String(), "venustech","tri", "1", "2", "50", "247825","2500222", "ES", "CME", "201709", "FUT", n,"0","0", "B1","VENUSTECH","TTDEV18O")
				if ordStatus.Status == "ok" {
					fmt.Println(ordStatus.Id)
					fmt.Printf("Order %s %s at %s Placed Successfully \n", ordStatus.Side, ordStatus.Symbol, ordStatus.Price)
				} else if ordStatus.Status == "rejected" {
					fmt.Printf("Order Rejected \n")
					if ordStatus.Reason != "" {
						fmt.Printf("Reason: %s\n", ordStatus.Reason)
					}
				}
			}
		case "20": //test multileg
			var u UnderlyingInstrumentGroup
			u.UnderlyingSecurityExchange="CME"
			u.UnderlyingSecurityType ="FUT"
			u.UnderlyingSymbol = "ES"
			u.UnderlyingMaturityMonthYear="201709"
			u.LegSide = "2"
			u.RatioQty="1"
			u.UnderlyingSecurityID= "00A0IR00ESZ"
			u.UnderlyingSecurityAltID = "ESU7"

			var u1 UnderlyingInstrumentGroup
			u1.UnderlyingSecurityExchange="CME"
			u1.UnderlyingSecurityType ="FUT"
			u1.UnderlyingSymbol = "ES"
			u1.UnderlyingMaturityMonthYear="201712"
			u1.LegSide = "1"
			u1.RatioQty="1"
			u1.UnderlyingSecurityID= "00A0LR00ESZ"
			u1.UnderlyingSecurityAltID = "ESZ7"

			var group []*UnderlyingInstrumentGroup
			group = append(group, &u)
			group= append(group, &u1)

			ordStatus:= TT_MultiLegNewOrder(xid.New().String(),"venustech","tri","1","2","5","-5500","","0","CME","Calendar",group,"VENUSTECH")
			if ordStatus.Status == "ok"{
				fmt.Println("Multi order placed successfully")
				fmt.Println(ordStatus.SecuritySubType)
				for _,u := range ordStatus.NoRelatedSymGroup{
					fmt.Printf("%s %s %s\n",u.UnderlyingSymbol,u.UnderlyingSecurityExchange, u.UnderlyingSecurityAltID)
				}
			}else{
				fmt.Println("Multileg order failed ")
				fmt.Printf("Reason: %s \n",ordStatus.Reason)
			}
		case "21":
			var group []*UnderlyingInstrumentGroup
			TT_MultiLegNewOrderAltid(xid.New().String(),"venustech","tri","1","2","5","2277","","0","CME","HOCL","CL:C1 HO-CL U7","Crack",group,"VENUSTECH")
		default:
			cmd := exec.Command("clear") //linux only
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
	fmt.Println("9) Get Fills")
	fmt.Println("99) Quit")
	fmt.Print("Action: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}
