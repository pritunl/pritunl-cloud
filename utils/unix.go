package utils

import (
	"crypto/sha512"
	"strconv"
)

const b64x24Chars = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func Base64x24(src []byte) (hash []byte) {
	if len(src) == 0 {
		return []byte{}
	}

	hashSize := (len(src) * 8) / 6
	if (len(src) % 6) != 0 {
		hashSize += 1
	}
	hash = make([]byte, hashSize)

	dst := hash
	for len(src) > 0 {
		switch len(src) {
		default:
			dst[0] = b64x24Chars[src[0]&0x3f]
			dst[1] = b64x24Chars[((src[0]>>6)|(src[1]<<2))&0x3f]
			dst[2] = b64x24Chars[((src[1]>>4)|(src[2]<<4))&0x3f]
			dst[3] = b64x24Chars[(src[2]>>2)&0x3f]
			src = src[3:]
			dst = dst[4:]
		case 2:
			dst[0] = b64x24Chars[src[0]&0x3f]
			dst[1] = b64x24Chars[((src[0]>>6)|(src[1]<<2))&0x3f]
			dst[2] = b64x24Chars[(src[1]>>4)&0x3f]
			src = src[2:]
			dst = dst[3:]
		case 1:
			dst[0] = b64x24Chars[src[0]&0x3f]
			dst[1] = b64x24Chars[(src[0]>>6)&0x3f]
			src = src[1:]
			dst = dst[2:]
		}
	}

	return
}

func GenerateShadow(passwd string) (output string, err error) {
	var i int
	rounds := 4096

	saltStr, err := RandStr(8)
	if err != nil {
		return
	}

	salt := []byte(saltStr)
	passwdByt := []byte(passwd)

	alternateHash := sha512.New()
	alternateHash.Write(passwdByt)
	alternateHash.Write(salt)
	alternateHash.Write(passwdByt)
	alernateSum := alternateHash.Sum(nil)

	aSeqHash := sha512.New()
	aSeqHash.Write(passwdByt)
	aSeqHash.Write(salt)
	for i = len(passwdByt); i > 64; i -= 64 {
		aSeqHash.Write(alernateSum)
	}
	aSeqHash.Write(alernateSum[0:i])

	for i = len(passwdByt); i > 0; i >>= 1 {
		if (i & 1) != 0 {
			aSeqHash.Write(alernateSum)
		} else {
			aSeqHash.Write(passwdByt)
		}
	}
	aSeqSum := aSeqHash.Sum(nil)

	pSeqHash := sha512.New()
	for i = 0; i < len(passwdByt); i++ {
		pSeqHash.Write(passwdByt)
	}
	pSeqSum := pSeqHash.Sum(nil)

	pSeq := make([]byte, 0, len(passwdByt))
	for i = len(passwdByt); i > 64; i -= 64 {
		pSeq = append(pSeq, pSeqSum...)
	}
	pSeq = append(pSeq, pSeqSum[0:i]...)

	sSeqHash := sha512.New()
	for i = 0; i < (16 + int(aSeqSum[0])); i++ {
		sSeqHash.Write(salt)
	}
	sSeqSum := sSeqHash.Sum(nil)

	sSeq := make([]byte, 0, len(salt))
	for i = len(salt); i > 64; i -= 64 {
		sSeq = append(sSeq, sSeqSum...)
	}
	sSeq = append(sSeq, sSeqSum[0:i]...)

	cSum := aSeqSum

	for i = 0; i < rounds; i++ {
		C := sha512.New()

		if (i & 1) != 0 {
			C.Write(pSeq)
		} else {
			C.Write(cSum)
		}

		if (i % 3) != 0 {
			C.Write(sSeq)
		}

		if (i % 7) != 0 {
			C.Write(pSeq)
		}

		if (i & 1) != 0 {
			C.Write(cSum)
		} else {
			C.Write(pSeq)
		}

		cSum = C.Sum(nil)
	}

	out := make([]byte, 0, 123)
	out = append(out, "$6$"...)
	out = append(out, []byte("rounds="+strconv.Itoa(rounds)+"$")...)
	out = append(out, salt...)
	out = append(out, '$')
	out = append(out, Base64x24([]byte{
		cSum[42], cSum[21], cSum[0],
		cSum[1], cSum[43], cSum[22],
		cSum[23], cSum[2], cSum[44],
		cSum[45], cSum[24], cSum[3],
		cSum[4], cSum[46], cSum[25],
		cSum[26], cSum[5], cSum[47],
		cSum[48], cSum[27], cSum[6],
		cSum[7], cSum[49], cSum[28],
		cSum[29], cSum[8], cSum[50],
		cSum[51], cSum[30], cSum[9],
		cSum[10], cSum[52], cSum[31],
		cSum[32], cSum[11], cSum[53],
		cSum[54], cSum[33], cSum[12],
		cSum[13], cSum[55], cSum[34],
		cSum[35], cSum[14], cSum[56],
		cSum[57], cSum[36], cSum[15],
		cSum[16], cSum[58], cSum[37],
		cSum[38], cSum[17], cSum[59],
		cSum[60], cSum[39], cSum[18],
		cSum[19], cSum[61], cSum[40],
		cSum[41], cSum[20], cSum[62],
		cSum[63],
	})...)

	aSeqHash.Reset()
	alternateHash.Reset()
	pSeqHash.Reset()
	for i = 0; i < len(aSeqSum); i++ {
		aSeqSum[i] = 0
	}
	for i = 0; i < len(alernateSum); i++ {
		alernateSum[i] = 0
	}
	for i = 0; i < len(pSeq); i++ {
		pSeq[i] = 0
	}

	output = string(out)

	return
}
