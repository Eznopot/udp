# UDP lib

## Introduction
Cette librairie est une librairie permettant de créer un client et un serveur UDP en Go.

Vous pouvez retrouver un exemple d'implémentation de cette lib dans le projet [Go-mutateur]("https://github.com/Eznopot/Go-mutateur")

## Specificité
Pour utiliser la librairie vous etes obliger d'utiliser le client et le serveur car il apparte un protocole spéacial permettant de gérer automatiquement la connection et déconnexion des clients.

## Fonctionement
### Serveur
Pour créer un serveur vous pouvez devez utiliser la fonction:
```Go
func CreateServer(port string, handler func(net.PacketConn, *net.Addr, udp.Packet)) sync.WaitGroup
```
La fonction `handler` est la fonction qui s'executera a chaque fois que votre serveur recevra un paquet. La fonction prend en parametre `net.PacketConn` qui est l'instance de votre serveur, `*net.Addr` qui est un pointeur sur l'addresse du client qui vous a envoyer un packet et `udp.Packet` qui est le model de donnée contenant les données transmis.

`CreateServer` renvoie `*sync.WaitGroup` qui permet de garder le controle de la go routine lancer par le serveur.

exemple de fonctionement:
```Go
import (
    udp_server "github.com/Eznopot/udp/server"
)

func handlerServer(udpServer net.PacketConn, addr *net.Addr, packet udp.Packet) {
	res, err := json.Marshal(packet)
	if err != nil {
		return
	}
	println(string(res))
}

func main() {
    wg := udp_server.CreateServer("8080", handlerServer)
    wg.Wait()
    udp_server.CloseServer()
    return
}
```
Pour renvoyer un packet a un ou plusieurs client vous pouvez utiliser les fonctions:
```Go
//envoie un message a tout les clients
func SendToAllClient(str, packetType string)

// envoie un message a tout les client sauf au client ayant l'address spécifié
func SendToAllExcludingItselfClient(str, packetType string, addr *net.Addr)

// envoie un message au client par sont index de connection (client le plus vieux 0)
func SendToClientByIndex(str, packetType string, index int)

// envoie un message au client via sont addresse
func SendToClientByAddress(str, packetType string, addr *net.Addr) 
```
L'argument str represente la string passée dans le message et le packetType est le type de packet que vous allez transferer (voir model plus bas)


Pour fermer un serveur proprement et close la connexion avec to les client vous pouvez utiliser la fonction
```Go
func CloseServer()
```

Vous pouvez récuperer la liste des ip des clients connecté a votre serveur en utilisant la fonction:
```Go
func GetAllClientInfo() []string 
```

Pour du debug ou autres vous aurez peut etre besoin d'afficher des information a la reception de vos requetes. Pour ca vous pourrez utiliser la fonction `SetLogger`:
```Go
func SetLogger(loggerFunc func(string))
```
Vous aurez juste a passer en parametre une fonction prenant en argument une string. Exemple:
```Go
import (
    udp_server "github.com/Eznopot/udp/server"
)

func handlerServer(udpServer net.PacketConn, addr *net.Addr, packet udp.Packet) {
	res, err := json.Marshal(packet)
	if err != nil {
		return
	}
    // do your stuff
}

func main() {
    wg := udp_server.CreateServer("8080", handlerServer)
	udp_server.SetLogger(func(s string) {
		println(s)
	})
    wg.Wait()
    udp_server.CloseServer()
    return
}
```

### Client
Pour créer un client et se connecter a un serveur et listen ce qu'il vous envois:
```Go 
func main() {
    var wg sync.WaitGroup
    udp_client.CreateConnection("localhost", "port")
    wg.Add(1);
	udp_client.Receive(&wg, func(packet udp.Packet) {
		// do your stuff here
	})
    wg.Wait()
    return
}
```
La fonction `Receive` prend en argument un `*sync.WaitGroup` qui recevra l'ordre de se terminer à la fermeture du serveur.

Pour fermer une connection proprement vous pouvez utiliser la fonction:
```Go
func CloseConnection()
```

Pour envoyer un message au serveur vous pouvez utiliser la fontion:
```Go
func SendToServer(str, packetType string)
```
## Model de donnée
La structure Packet qui est systematiquement transferer contient les elements `Data` et `Type`.
Lors de la premiere connection d'un client un message ayant le type `handshake` est automoatiquement envoyer sinifiant au serveur de rajouter ce client dans la liste de tout les clients connecté.
Lors de la deconnexion d'un client ou du serveur un message avec le Type `close` est automatiquement envoyer, permettant de fermé la connection proprement avec le client ou le serveur.
```Go
type Packet struct {
	Data string `json:"data"`
	Type string `json:"type"`
}
```