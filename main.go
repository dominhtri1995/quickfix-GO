package main  
import (
"fmt"
"os"
"bufio"
)
func main() {
	
	StartQuickFix()
Loop:
	for {
		action, err := QueryAction()
		if err != nil {
			break
		}
//////////		
		switch action {
		case "1":
			QueryPAndL("5","TTORDFA222222")
		case "2":
			QueryNewOrderSingle("28","venustech1","1","2","500","4570","BZ","CME","201709")
		case "3":
			QueryWorkingOrder("venustech")
		case "4":
			//quit
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
	fmt.Println("1) Enter Order")
	fmt.Println("2) Cancel Order")
	fmt.Println("3) Request Market Test")
	fmt.Println("4) Quit")
	fmt.Print("Action: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}