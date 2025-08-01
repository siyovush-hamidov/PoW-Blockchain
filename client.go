package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	bc "github.com/siyovush-hamidov/Blockchain/blockchain"
	nt "github.com/siyovush-hamidov/Blockchain/network"
)

const (
	ADD_BLOCK = iota + 1
	// Q: iota
	ADD_TRNSX
	GET_BLOCK
	GET_LHASH
	GET_BLNCE
	GET_CSIZE
)

var (
	Addresses []string
	User *bc.User
)

func init() {
	if len(os.Args) < 2 {
		// Q: что за os.Args?
		panic("failed: len(os.Args < 2)")
	}
	var (
		addrStr = ""
		userNewStr = ""
		userLoadStr = ""
	)
	// Q: Зачем создавать 2 var, если можно всё поставить в один?
	var (
		addrExist = false
		userNewExist = false
		userLoadExist = false
	)
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case strings.HasPrefix(arg, "-loadaddr:"):
			// Q: Что тут сейчас произошло?
			addrStr = strings.Replace(arg, "-loadaddr:", "", 1)
			addrExist = true
		case strings.HasPrefix(arg, "-newuser:"):
			userNewStr = strings.Replace(arg, "-newuser:", "", 1)
			userNewExist = true
		case strings.HasPrefix(arg, "-loaduser:"):
			userLoadStr = strings.Replace(arg, "-loaduser:", "", 1)
			userLoadExist = true
		}
	}
	if !(userNewExist || userLoadExist || !addrExist) {
		panic("failed: !(userNewExist || userLoadExist || !addrExist)")
	}

	err := json.Unmarshal([]byte(readFile(addrStr)), &Addresses)
	if err != nil {
		panic("failed: load addresses")
	}
	if len(Addresses) == 0 {
		panic("failed: len(Addresses) == 0")
	}
	if userNewExist {
		User = userNew(userNewStr)
	}
	if userLoadExist {
		User = userLoad(userLoadStr)
	}
	if User == nil {
		panic("failed: load user")
	}

    fmt.Println("User address:", User.Address())
    fmt.Println("Loaded addresses:", Addresses)
}

func readFile(filename string) string {
	data, err := os.ReadFile(filename)
	// Q: Что за ioutil? inout output util?
	if err != nil {
		return ""
	}
	return string(data)
}

func userNew(filename string) *bc.User {
	user := bc.NewUser()
	if user == nil {
		return nil
	}
	err := writeFile(filename, user.Purse())
	if err != nil {
		return nil
	}
	return user
}

func userLoad(filename string) *bc.User {
	priv := readFile(filename)
	if priv == "" {
		return nil
	}	
	user := bc.LoadUser(priv)
	if user == nil {
		return nil
	}
	return user
}

func writeFile(filename string, data string) error {
	return os.WriteFile(filename, []byte(data), 0644)
}

func handleClient() {
	var (
		message string
		splited []string
	)
	// Q: Где сохраняются переменные внутри функции, глобальные и переменные, для который динамически выделена память?
	for {
		message = inputString("> ")
		// Q: Что за синтаксис? inputString(begin string) и туда мы пишем "> ". Что это значит?
		splited = strings.Split(message, " ")
		switch splited[0] {
		case "/exit":
			os.Exit(0)
			// Q: Что значит 0 в exit(0) в языках программироавания? 
		case "/user":
			if len(splited) < 2 {
				fmt.Println("failed: len(user) < 2\n")
				continue
			}
			switch splited[1] {
			case "address":
				userAddress()
			case "purse":
				userPurse()
			case "balance":
				userBalance()
			default:
				fmt.Println("command undefined\n")
			}
		case "/chain":
			if len(splited) < 2 {
				fmt.Println("Failed: len(chain) < 2\n")
				continue
			}
			switch splited[1] {
			case "print":
				chainPrint()
			case "tx":
				chainTX(splited[1:])
			case "balance":
				chainBalance(splited[1:])
			case "block":
				chainBlock(splited[1:])
			case "size":
				chainSize()
			default:
				fmt.Println("command undefined\n")
			}
		default:
			fmt.Println("command undefined\n")
		}	
	}
}

func inputString(begin string) string {
	fmt.Print(begin)
	// Q: fmt.Print vs fmt.Println
	msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.Replace(msg, "\n", "", 1)
}

func userAddress() {
	fmt.Println("Address:", User.Address(), "\n")
}

func userPurse() {
	fmt.Println("Purse:", User.Purse(), "\n")
}

func userBalance() {
	printBalance(User.Address())
}

func chainPrint() {
	for i := 0; ; i++ {
		// Q: Что за синтаксис?
		res := nt.Send(Addresses[0], &nt.Package{
			Option: GET_BLOCK,
			Data: fmt.Sprintf("%d", i),
		})
		if res == nil || res.Data == ""{
			break
		}
		fmt.Printf("[%d] => %s\n", i+1, res.Data)
	}
	fmt.Println()
}
func chainTX(splited []string) {
    if len(splited) != 3 {
        fmt.Println("failed: len(splited) != 3\n")
        return
    }
    num, err := strconv.Atoi(splited[2])
    if err != nil {
        fmt.Println("failed: strconv.Atoi(num)\n")
        return
    }
    for _, addr := range Addresses {
        fmt.Println("Sending GET_LHASH to", addr)
        res := nt.Send(addr, &nt.Package{
            Option: GET_LHASH,
        })
        if res == nil {
            fmt.Println("No response from", addr)
            continue
        }
        tx := bc.NewTransaction(User, bc.Base64Decode(res.Data), splited[1], uint64(num))
        fmt.Println("Sending ADD_TRNSX to", addr)
        res = nt.Send(addr, &nt.Package{
            Option: ADD_TRNSX,
            Data: bc.SerializeTX(tx),
        })
        if res == nil {
            fmt.Println("No response from", addr)
            continue
        }
        if res.Data == "ok" {
            fmt.Printf("ok: (%s)\n", addr)
        } else {
            fmt.Printf("fail: (%s)\n", addr)
        }
    }
    fmt.Println()
} 
// Q: strconv.Atoi

func chainBalance(splited []string) {
	if len(splited) != 2 {
		fmt.Println("fail: len(splited) != 2\n")
		return
	}
	printBalance(splited[1])
}

func chainBlock(splited []string) {
	if len(splited) != 2 {
		fmt.Println("failed: len(splited) != 2\n")
		return
	}
	num, err := strconv.Atoi(splited[1])
	if err != nil {
		fmt.Println("failed: strconv.Atoi(num)\n")
		return
	}
	res := nt.Send(Addresses[0], &nt.Package{
		Option: GET_BLOCK,
		Data: fmt.Sprintf("%d", num - 1),
		// Q: fmt.Sprintf()
	})
	if res == nil || res.Data == "" {
		fmt.Println("failed: getBlock\n")
		return
	}
	fmt.Printf("[%d] => %s\n", num, res.Data)
}

func chainSize() {
	res := nt.Send(Addresses[0], &nt.Package{
		Option: GET_CSIZE,
	})
	if res == nil || res.Data == "" {
		fmt.Println("failed: getSize\n")
		return
	}
	fmt.Printf("Size: %s blocks\n\n", res.Data)
}
func printBalance(useraddr string) {
    for _, addr := range Addresses {
        fmt.Println("Sending GET_BLNCE to", addr)
        res := nt.Send(addr, &nt.Package{
            Option: GET_BLNCE,
            Data: useraddr,
        })
        if res == nil {
            fmt.Println("No response from", addr)
            continue
        }
        fmt.Printf("Balance (%s): %s coins\n", addr, res.Data)
    }
    fmt.Println()
}

func main() {
	handleClient()
}