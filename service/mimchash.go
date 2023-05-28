package service
import (
	"fmt"
	"math/big"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"

)

func Mimchash(str string) string {
	ds := mimc.NewMiMC()
	bigIntValue := new(big.Int)
	_, success := bigIntValue.SetString(str, 10) // 从十进制字符串创建 big.Int
	if !success {
		fmt.Println("Error converting string to big.Int")
		return ""
	}
	// 将 big.Int 转换为字节切片
	data := bigIntValue.Bytes()

	// 将数据写入哈希实例
	n, err := ds.Write(data)
	if err != nil {
		fmt.Println("error:", err.Error())
	}
	fmt.Println("Bytes written:", n)

	// 计算哈希值
	hashValue := ds.Sum(nil)
	return string(hashValue)
}