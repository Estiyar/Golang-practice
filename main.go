func goroutineHello() {  
fmt.Println("goroutine hello")  
}  
func main() {  
fmt.Println("main hello")  
go goroutineHello()  
}
