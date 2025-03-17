package utils

import (
	"fmt"
	"math"
)

func UintToInt32Safe(val uint) (int32, error) {
	if val > math.MaxInt32 {
		return 0, fmt.Errorf("value %d exceeds maximum int32 value", val)
	}

	return int32(val), nil
}

func UintPtrToInt32PtrSafe(uintPtr *uint) (*int32, error) {
	if uintPtr == nil {
		return nil, nil
	}

	int32Val, err := UintToInt32Safe(*uintPtr)
	if err != nil {
		return nil, err
	}

	return &int32Val, nil
}

func Int32ToUintSafe(val int32) (uint, error) {
	if val < 0 {
		return 0, fmt.Errorf("value %d is negative", val)
	}

	return uint(val), nil
}

func Int32PtrToUintSafe(int32Ptr *int32) (uint, error) {
	if int32Ptr == nil {
		return 0, nil
	}

	return Int32ToUintSafe(*int32Ptr)
}

func Int32PtrToUintPtrSafe(int32Ptr *int32) (*uint, error) {
	if int32Ptr == nil {
		return nil, nil
	}

	uintVal, err := Int32ToUintSafe(*int32Ptr)
	if err != nil {
		return nil, err
	}

	return &uintVal, nil
}
