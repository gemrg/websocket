package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dgrr/websocket"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type Broadcaster struct {
	cs  map[uint64]*websocket.Conn
}

// it is concurrently safe
func (b *Broadcaster) OnOpen(c *websocket.Conn) {
	b.cs[c.ID()] = c
}

func (b *Broadcaster) OnClose(c *websocket.Conn, err error) {
	if err != nil {
		log.Printf("%d closed with error: %s\n", c.ID(), err)
	} else {
		log.Printf("%d closed the connection\n", c.ID())
	}

	delete(b.cs, c.ID())
}

func (b *Broadcaster) Start() {
	for i := 0; ; i++ {
		for _, nc := range b.cs {
			fmt.Fprintf(nc, "Sending message number %d\n", i)
		}

		time.Sleep(time.Second)
	}
}

func main() {
	b := &Broadcaster{
		cs: make(map[uint64]*websocket.Conn),
	}

	wServer := websocket.Server{}
	wServer.HandleOpen(b.OnOpen)
	wServer.HandleClose(b.OnClose)

	router := router.New()
	router.GET("/", rootHandler)
	router.GET("/ws", wServer.Upgrade)

	server := fasthttp.Server{
		Handler: router.Handler,
	}

	go b.Start()

	server.ListenAndServe(":8080")
}

func rootHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/html")
	fmt.Fprintln(ctx, `<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8"/>
    <title>Sample of websocket with Golang</title>
  </head>
  <body>
		<div id="text"></div>
    <script>
      var ws = new WebSocket("ws://localhost:8080/ws");
      ws.onmessage = function(e) {
				var d = document.createElement("div");
        d.innerHTML = e.data;
				ws.send(e.data);
        document.getElementById("text").appendChild(d);
      }
			ws.onclose = function(e){
				var d = document.createElement("div");
				d.innerHTML = "CLOSED";
        document.getElementById("text").appendChild(d);
			}
    </script>
  </body>
</html>`)
}
