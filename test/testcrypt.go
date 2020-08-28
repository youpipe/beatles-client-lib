package main

import (
	"encoding/hex"
	"fmt"
	"github.com/kprc/libeth/util"
)

func main() {
	plainTxt := "123456789123456789123456789"

	key, err := hex.DecodeString("d971617ee36e0266740b9ac8badc954c875706e921ac67531d152bf1cdf1433e")
	if err != nil {
		fmt.Println(err)
		return
	}

	stream, iv, err := util.NewEncStream(key)
	ciphertxt := util.Encrypt2(stream, []byte(plainTxt))

	fmt.Println(hex.EncodeToString(ciphertxt), len(ciphertxt))

	stream1, err1 := util.NewDecStreamWithIv(key, iv)
	if err1 != nil {
		fmt.Println(err1)
		return
	}

	plainTxt1 := util.Decrypt2(stream1, ciphertxt)

	fmt.Println(string(plainTxt1), len(plainTxt1))

}
