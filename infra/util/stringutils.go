package util

import "hash/crc32"

func HashCode (s *string) int {
	v:=int(crc32.ChecksumIEEE([]byte(*s)))
	if v>=0 {
		return v
	}
	if v<=0 {
		return -v
	}
	return 0
}
