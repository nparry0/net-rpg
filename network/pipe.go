package network

/****************************************
 * Pipe implementation by Maxim Khitrov
 * http://play.golang.org/p/P8hzE0zt12
 * Modified for use here
 ****************************************/

//type Type int

func NewPipe() (from chan<- GameMsg, to <-chan GameMsg) {
	f := make(chan GameMsg, 1)
	t := make(chan GameMsg, 1)
	go pipe(f, t)
	return f, t
}

func pipe(from <-chan GameMsg, to chan<- GameMsg) {
	var i, cnt int
	var zero GameMsg
	defer close(to)

	ring := make([]GameMsg, 2)
	resize := func(size int) {
		temp := make([]GameMsg, size)
		if i+cnt <= len(ring) {
			copy(temp, ring[i:i+cnt])
		} else {
			n := copy(temp, ring[i:])
			copy(temp[n:], ring[:cnt-n])
		}
		i, ring = 0, temp
	}
	pop := func() {
		ring[i] = zero
		i, cnt = (i+1)%len(ring), cnt-1
		if cnt < len(ring)/2 && len(ring) > 2 {
			resize(len(ring) / 2)
		}
	}

recv:
	for v := range from {
		select {
		case to <- v:
			continue
		default:
			i, cnt, ring[0] = 0, 1, v
		}
		for cnt > 0 {
			select {
			case to <- ring[i]:
				pop()
			case v, ok := <-from:
				if !ok {
					break recv
				} else if cnt == len(ring) {
					resize(len(ring) * 2)
				}
				ring[(i+cnt)%len(ring)] = v
				cnt++
			}
		}
	}
	for cnt > 0 {
		to <- ring[i]
		pop()
	}
}

/*
func main() {
	send, recv := NewPipe()
	for i := Type(0); i < 20; i++ {
		send <- i
	}
	for i := 0; i < 10; i++ {
		fmt.Println(<-recv)
	}
	for i := Type(20); i < 40; i++ {
		send <- i
	}
	close(send)
	for v := range recv {
		fmt.Println(v)
	}
}
*/
