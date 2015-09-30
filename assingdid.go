package getrelayv2

//import "io/util"

/*
var (
	header_prefix string = "POST /agent/relay/get HTTP/1.1\r\nHOST: api.dch.dlink.com:443\r\n" +
			"Content-Type: application/x-www-form-urlencoded\r\n"
	header_prefix_clen string = "Content-Length:"
	header_prefix_default_len string = "171"
	header_prefix_nextline string = "\r\n"
	header_prefix_close string = "Connection: close\r\n\r\n"
	header_default_did string = "did=48e726f9ea31bec30c7486ddc733365a"
	header_default_p string = "p=qZA9XmwJ9b8Zzo06qu0zfxrhm6Rs9iY3Lm1sNeH5PAjmVGNovBG6coHi7aWGFo%2BmQKTJld%2BqahuOtFEh%2BTpfHA%3D%3D"
	header_default_iv string = "iv=sT%2B0mdtrwqrsdasphTx3Fw%3D%3D"
	default_msg string = header_prefix + header_prefix_clen + header_prefix_default_len + header_prefix_nextline
		+ header_prefix_close + header_default_did + "&" + header_default_p + "&" + header_default_iv + header_prefix_nextline
)
*/
/*https://api.dch.dlink.com/agent/v2/relay/get*/
/*
POST /connect HTTP/1.1\r\n
Content-Type: text\r\n
\r\n
"did":"1aa2e235723dd6a300381edc551aab3c"\r\n
"hash":"bff84d880edabb2c1598c537906d04d9"\r\n
*/
/*
func AssingDid(did string) bool {
	url := "api.dch.dlink.com:443"
	var ret bool = false
	tcpaddr, err := net.ResolveTCPAddr("tcp", url)
	//tcpaddr, err := net.ResolveTCPAddr("tcp", "52.76.5.88:80")
	log.Println("ResolveTCPAddr = ", tcpaddr.IP)

	if err != nil {
		log.Println("[assigndid]error", err, " url=", url)
		return false
	}

	conn, err := net.DialTCP("tcp", nil, tcpaddr) //Connect to server

	if err != nil {
		log.Println("[assigndid] connect error", err, "url = ", url)
		return false
	}

	conn.Write([]byte(default_msg))

	return ret
}
*/
