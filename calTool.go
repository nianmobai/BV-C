package main

import (
	"math/rand"
)

const QuickSort = 0
const SelectSort = 1

// 费马定理判断质数
func ispremeFetmats(N int) (bool, bool) {
	if N < 3 {
		if N <= 0 {
			return false, false
		} else {
			return false, true
		}
	} else {
		for i := 0; i < 20; i++ {
			a := rand.Intn(N-2) + 1
			result := QuickExponentiation(a, N-1) % N
			if result != 1 {
				return false, true
			}
		}
		return true, true
	}
}

// 快速幂
func QuickExponentiation(a int, b int) int {
	result := a
	for b != 0 {
		if b&1 == 1 {
			result *= a
		} else {
			result *= result
		}
		b >>= 1
	}
	return result
}

// 米勒拉宾容易爆int，先不做
func isperemMillerRabin(N int) (bool, bool) {
	return false, false
}

// 朴素的筛质数方法
func isperemSimple(N int) (bool, bool) {
	if N < 0 {
		return false, false
	}
	if N%2 == 0 {
		return false, true
	}
	for i := 3; i*i < N; i += 2 {
		if N%i == 0 {
			return false, true
		}
	}
	return true, true
}

// EuclideanSieve 欧式筛
func EuclideanSieve(N int) ([]int, bool) {
	if N <= 1 {
		return []int{}, false
	}
	var result []int
	num := make([]int, N)
	for i := range num {
		num[i] = 0
	}
	result = append(result, 2)
	for i := 3; i < N; i += 2 {
		if num[i] == 0 {
			result = append(result, i)
			for j := i; i*j < N; j++ {
				num[i*j] = 1
			}
		}
	}
	return result, true
}

// EulerSieve 质数筛
func EulerSieve(N int) ([]int, bool) {
	if N <= 1 {
		return []int{}, false
	}
	var result []int
	num := make([]int, N+1)
	for i := range num {
		num[i] = 0
	}
	result = append(result, 2)
	for i := 3; i < N; i += 2 {
		if num[i] == 0 {
			result = append(result, i)
		}
		for j := 0; j < len(result) && (i*result[j]) < N; j++ {
			num[i*result[j]] = 1
			if i%result[j] == 0 {
				break
			}
		}
	}
	return result, true
}

// ///////////////////////
// 排序
//
// ////////////////
// 排序工具
func sortTool(arr []int, button int) []int {
	var result []int
	switch button {
	case QuickSort:
		result = sortQuick(arr)
	case SelectSort:
		result = sortSelect(arr)
	}
	return result
}

// 快速排序
func sortQuick(arr []int) []int {
	end := len(arr) - 1
	base := arr[end]
	var left []int
	var right []int
	for _, val := range arr {
		if base <= val {
			left = append(left, val)
		} else {
			right = append(right, val)
		}
	}
	left = sortQuick(left)
	right = sortQuick(right)
	result := append([]int{}, left...)
	result = append(result, base)
	result = append(result, right...)
	return result
}

// 冒泡排序
func sortSelect(arr []int) []int {
	length := len(arr)
	for i := 0; i < length; i++ {
		var minVal int
		var minTag int
		minTag = i
		minVal = arr[i]
		for j := i; j < length; j++ {
			if arr[j] < minVal {
				minTag = j
				minVal = arr[j]
			}
		}
		arr[minTag], arr[i] = arr[i], arr[minTag] //交换值
	}
	return arr
}

// 归并排序
func sortMerge(arr []int) []int {
	if len(arr) == 0 || len(arr) == 1 {
		return arr
	}
	var result []int
	length := len(arr)
	middle := length / 2
	var left, right []int
	left = sortSelect(arr[:middle])
	right = sortSelect(arr[middle:])
	i := 0
	j := 0
	for i < len(left) || j < len(right) {
		if i < len(left) && j < len(right) {
			if left[i] < right[j] {
				result = append(result, left[i])
				i++
			} else {
				result = append(result, right[j])
				j++
			}
		} else if i >= len(left) { //左边超限
			result = append(result, right[j])
			j++
		} else { //右侧超限
			result = append(result, left[i])
			i++
		}
	}
	return result
}

// 堆排序
func sortHeap(arr []int) []int {
	var result []int

	return result
}
