package main

import (
	"fmt"
	"time"
)
func main(){

	//slice := []int{1 , 3, 5}
	//fmt.Println(cap(slice))
	//
	//new_slice := append(slice , 1 , 2 , 3)
	//fmt.Println(1,&new_slice[0])
	//
	//new_slice[0] = 999
	//fmt.Println(2,&slice[0])
	//fmt.Println(3,&new_slice[0])
	//
	// new_slice2 := make([]int , 3 )
	// copy(new_slice2 , slice)
	// slice[0] = 888
	// fmt.Println(4,&new_slice2[0])
	//
	//fmt.Println("before goroutine",runtime.NumGoroutine())
	//
	//a := make(chan int)
	//for i:=0;i<3;i++{
	//
	//	 go func( j int){
	//		 a <- j
	//	 }(i)
	//
	//}
	//fmt.Println("after goroutine",runtime.NumGoroutine())
	//
	//for i:=0;i<3;i++{
	//	res := <- a
	//	fmt.Println(res)
	//}

	//fmt.Println("before goroutine",runtime.NumGoroutine())
	//
	//
	//
	//	go func( ){
	//		for i:= 0 ; i<3;i++{
	//			fmt.Println(i)
	//		}
	//
	//	}()
	//
	//
	//fmt.Println("after goroutine",runtime.NumGoroutine())
	//go printOneNumber()
	//fmt.Println("----")
	//go number(2)
	//var input string
	//fmt.Scanln(&input)

	//var networkBootstrapped chan bool


	//fmt.Println("before goroutine",runtime.NumGoroutine())
	//
	//potentialPeerList := []int{1,2,3,4,5,6,7,8,9,10}
	//go announceToNetwork(potentialPeerList)
	//fmt.Println("after goroutine",runtime.NumGoroutine())



	//a := make(chan int)
	//
	//go func() {
	//	a <- 2
	//}()
	//
	//res := <- a
	//fmt.Println(res)

	a := []int{}
	fmt.Println(len(a))






}

func announceToNetwork(List []int){

	for _,value := range List{
		fmt.Println(value)
	}

}

func callRegisterPeerRPC( number int) {
	 c := make(chan int )
		go func() {
			c <- RegisterBlock(number)
		}()

	 res := <-c
	 fmt.Println(res)
	}

func RegisterBlock(value int ) (res int) {
	value += 1

	return
}



func number( max int) {
	for j := 0 ; j <= max ;j++{
		fmt.Println(j)
		time.Sleep(time.Second)

	}
}

func printOneNumber(){
	fmt.Println(2)
}