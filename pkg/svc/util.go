package svc

import "github.com/lithammer/go-jump-consistent-hash"

const bucketSize = 1024
func hash(key string)int{
    return int(jump.HashString(key, bucketSize, jump.NewCRC64()))
}

func BinarySearch(key int, arr []int, startIndex,orgLen int)int{
    if 0 == orgLen{
        return -1
    }

    arrLen := len(arr)
    if 0 == arrLen{
        return 0
    }

    if 1 == arrLen{
        if key <= arr[0]{
            return startIndex
        }

        // 检测是否存在下一个元素
        if startIndex < orgLen - 1{
            return startIndex + 1
        }else { //已经是最后一个元素
            return 0
        }
    }

    mid := arrLen/2
    if key >= arr[mid]{
        return BinarySearch(key, arr[mid:], startIndex + mid, orgLen)
    }

    return BinarySearch(key, arr[:mid] ,startIndex, orgLen)
}
