package progress

import (
	"context"
	"strconv"
	"testing"
	"time"
)

func startHub(t *testing.T) (*Hub, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	h := NewHub()
	go h.Run(ctx)
	t.Cleanup(cancel)
	return h, cancel
}

// recv espera um evento por até 1 s, para um teste com defeito não travar.
func recv(t *testing.T, c <-chan Event) Event {
	t.Helper()
	select {
	case ev, ok := <-c:
		if !ok {
			t.Fatal("channel fechado")
		}
		return ev
	case <-time.After(time.Second):
		t.Fatal("nenhum evento em 1 s")
	}
	return Event{}
}

func TestTodosOsInscritosRecebem(t *testing.T) {
	h, _ := startHub(t)
	a, b := h.Subscribe(), h.Subscribe()

	h.Publish(Event{ID: "1", Name: "foto.jpg", Done: true})

	for _, c := range []chan Event{a, b} {
		if ev := recv(t, c); ev.ID != "1" || !ev.Done {
			t.Errorf("evento inesperado: %+v", ev)
		}
	}
}

func TestUnsubscribeFechaOChannel(t *testing.T) {
	h, _ := startHub(t)
	c := h.Subscribe()
	h.Unsubscribe(c)
	select {
	case _, ok := <-c:
		if ok {
			t.Fatal("esperava channel fechado")
		}
	case <-time.After(time.Second):
		t.Fatal("Unsubscribe não fechou o channel")
	}
}

// Um inscrito que não lê não pode travar os uploads: o hub descarta os
// eventos dele e segue entregando para os outros.
func TestInscritoLentoNaoTravaNinguem(t *testing.T) {
	h, _ := startHub(t)
	_ = h.Subscribe() // nunca lê
	rapido := h.Subscribe()

	got := make(chan int)
	go func() {
		n := 0
		for ev := range rapido {
			n++
			if ev.Done {
				got <- n
				return
			}
		}
	}()

	publicado := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			h.Publish(Event{ID: "x", Bytes: int64(i)})
		}
		h.Publish(Event{ID: "x", Done: true})
		close(publicado)
	}()

	select {
	case <-publicado:
	case <-time.After(2 * time.Second):
		t.Fatal("Publish travou por causa do inscrito lento")
	}
	select {
	case n := <-got:
		if n < 1 {
			t.Error("o inscrito rápido não recebeu nada")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("o inscrito rápido não recebeu o evento final")
	}
}

func TestPararOHubFechaInscritosESoltaPublish(t *testing.T) {
	h, cancel := startHub(t)
	c := h.Subscribe()
	cancel()

	select {
	case _, ok := <-c:
		if ok {
			t.Fatal("esperava channel fechado")
		}
	case <-time.After(time.Second):
		t.Fatal("o hub parou mas não fechou o channel")
	}

	done := make(chan struct{})
	go func() {
		h.Publish(Event{Done: true}) // com o hub parado não pode travar
		h.Unsubscribe(h.Subscribe())
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Publish/Subscribe travaram com o hub parado")
	}
}

// Eventos que já estavam na fila quando o hub parou ainda são entregues.
// Sem o drain, o select do Run sorteia entre "parar" e "entregar" e alguns
// destes 50 se perderiam (este teste falha na maioria das vezes sem ele).
func TestPararOHubEntregaOQueJaEstavaNaFila(t *testing.T) {
	for rodada := 0; rodada < 20; rodada++ {
		h, cancel := startHub(t)
		c := h.Subscribe()
		for i := 0; i < 50; i++ { // 50 cabem no buffer do inscrito (64)
			h.Publish(Event{ID: strconv.Itoa(i), Done: true})
		}
		cancel()

		n := 0
		for range c { // range termina quando o hub fecha o channel
			n++
		}
		if n != 50 {
			t.Fatalf("rodada %d: o inscrito recebeu %d de 50 eventos", rodada, n)
		}
	}
}
