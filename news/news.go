package news

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
)

func taskWithParams(a int, b string) {
	fmt.Println(a, b)
}

func ad() {

	s1 := gocron.NewScheduler(time.UTC)
	s1.Every(1).Second().Do(taskWithParams, 1, "hello")
	s1.StartAsync()

	fmt.Println("hier")

	time.Sleep(6000 * time.Millisecond)

}
