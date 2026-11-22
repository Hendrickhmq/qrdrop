// Pacote netutil descobre o endereço da máquina na rede local.
package netutil

import (
	"errors"
	"net"
)

// LocalIP devolve o IPv4 que a máquina usa para falar com a rede local.
//
// O truque: "conectar" um socket UDP a um IP externo não envia nenhum pacote,
// mas faz o sistema operacional escolher a interface de saída. O endereço
// local desse socket é o IP que os outros aparelhos da rede enxergam.
func LocalIP() (net.IP, error) {
	conn, err := net.Dial("udp4", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok && !addr.IP.IsLoopback() {
			return addr.IP, nil
		}
	}
	// Sem rota padrão (rede sem internet): procura nas interfaces.
	return firstPrivateIP()
}

func firstPrivateIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		if ip := ipnet.IP.To4(); ip != nil && ip.IsPrivate() {
			return ip, nil
		}
	}
	return nil, errors.New("nenhum IPv4 de rede local encontrado")
}
