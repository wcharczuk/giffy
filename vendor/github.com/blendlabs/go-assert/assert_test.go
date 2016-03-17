package assert

import (
	"strings"
	"testing"
	"time"
)

func TestEmpty(t *testing.T) {
	a := Empty()
	if a.NonFatal().True(false, "this should fail") {
		t.Error("NonFatal true(false) didn't fail.")
	}
	if !a.NonFatal().True(true, "this should pass") {
		t.Error("NonFatal true(true) didn't pass.")
	}
}

func TestIsZero(t *testing.T) {

	zero_short := int16(0)
	if !isZero(zero_short) {
		t.Error("isZero failed")
	}
	not_zero_short := int16(3)
	if isZero(not_zero_short) {
		t.Error("isZero failed")
	}

	zero := 0
	if !isZero(zero) {
		t.Error("isZero failed")
	}
	not_zero := 3
	if isZero(not_zero) {
		t.Error("isZero failed")
	}

	zero_float64 := 0.0
	if !isZero(zero_float64) {
		t.Error("isZero failed")
	}
	not_zero_float64 := 3.14
	if isZero(not_zero_float64) {
		t.Error("isZero failed")
	}

	zero_float32 := float32(0.0)
	if !isZero(zero_float32) {
		t.Error("isZero failed")
	}
	not_zero_float32 := float32(3.14)
	if isZero(not_zero_float32) {
		t.Error("isZero failed")
	}
}

func TestGetLength(t *testing.T) {
	empty_string := ""
	l := getLength(empty_string)
	if l != 0 {
		t.Errorf("getLength incorrect.")
	}

	not_empty_string := "foo"
	l = getLength(not_empty_string)
	if l != 3 {
		t.Errorf("getLength incorrect.")
	}

	empty_array := []int{}
	l = getLength(empty_array)
	if l != 0 {
		t.Errorf("getLength incorrect.")
	}

	not_empty_array := []int{1, 2, 3}
	l = getLength(not_empty_array)
	if l != 3 {
		t.Errorf("getLength incorrect.")
	}

	empty_map := map[string]int{}
	l = getLength(empty_map)
	if l != 0 {
		t.Errorf("getLength incorrect.")
	}

	not_empty_map := map[string]int{"foo": 1, "bar": 2, "baz": 3}
	l = getLength(not_empty_map)
	if l != 3 {
		t.Errorf("getLength incorrect.")
	}
}

type myNestedStruct struct {
	Id   int
	Name string
}

type myTestStruct struct {
	Id          int
	Name        string
	SingleValue float32
	DoubleValue float64
	Timestamp   time.Time
	Struct      myNestedStruct

	IdPtr     *int
	NamePptr  *string
	StructPtr *myNestedStruct

	Slice    []myNestedStruct
	SlicePtr *[]myNestedStruct
}

func createTestStruct() myTestStruct {

	test_int := 1
	test_name := "test struct"

	nested_a := myNestedStruct{1, "A"}
	nested_b := myNestedStruct{1, "B"}
	nested_c := myNestedStruct{1, "C"}

	test_struct := myTestStruct{}
	test_struct.Id = test_int
	test_struct.Name = test_name
	test_struct.SingleValue = float32(3.14)
	test_struct.DoubleValue = 6.28
	test_struct.Timestamp = time.Now()
	test_struct.Struct = nested_a

	test_struct.IdPtr = &test_int
	test_struct.NamePptr = &test_name
	test_struct.StructPtr = &nested_b

	test_struct.Slice = []myNestedStruct{nested_a, nested_b, nested_c}
	test_struct.SlicePtr = &test_struct.Slice
	return test_struct

}

func TestStructsAreEqual(t *testing.T) {
	test_struct_a := createTestStruct()
	test_struct_b := createTestStruct()
	test_struct_b.Name = "not test struct"

	if did_fail, _ := shouldBeEqual(test_struct_a, test_struct_a); did_fail {
		t.Error("shouldBeEqual Failed.")
		t.FailNow()
	}

	if did_fail, _ := shouldBeEqual(test_struct_a, test_struct_b); !did_fail {
		t.Error("shouldBeEqual Failed.")
		t.FailNow()
	}
}

func TestShouldBeEqual(t *testing.T) {
	byte_a := byte('a')
	byte_b := byte('b')

	if did_fail, _ := shouldBeEqual(byte_a, byte_a); did_fail {
		t.Error("shouldBeEqual Failed.")
		t.FailNow()
	}
	if did_fail, _ := shouldBeEqual(byte_a, byte_b); !did_fail {
		t.Error("shouldBeEqual Failed.")
		t.FailNow()
	}

	string_a := "test"
	string_b := "not test"

	if did_fail, _ := shouldBeEqual(string_a, string_a); did_fail {
		t.Error("shouldBeEqual Equal Failed.")
		t.FailNow()
	}
	if did_fail, _ := shouldBeEqual(string_a, string_b); !did_fail {
		t.Error("shouldBeEqual Failed.")
		t.FailNow()
	}

	int_a := 1
	int_b := 2

	if did_fail, _ := shouldBeEqual(int_a, int_a); did_fail {
		t.Error("shouldBeEqual Equal Failed.")
		t.FailNow()
	}
	if did_fail, _ := shouldBeEqual(int_a, int_b); !did_fail {
		t.Error("shouldBeEqual Failed.")
		t.FailNow()
	}

	float32_a := float32(3.14)
	float32_b := float32(6.28)

	if did_fail, _ := shouldBeEqual(float32_a, float32_a); did_fail {
		t.Error("shouldBeEqual Equal Failed.")
		t.FailNow()
	}
	if did_fail, _ := shouldBeEqual(float32_a, float32_b); !did_fail {
		t.Error("shouldBeEqual Failed.")
		t.FailNow()
	}

	float_a := 3.14
	float_b := 6.28

	if did_fail, _ := shouldBeEqual(float_a, float_a); did_fail {
		t.Error("shouldBeEqual Equal Failed.")
		t.FailNow()
	}
	if did_fail, _ := shouldBeEqual(float_a, float_b); !did_fail {
		t.Error("shouldBeEqual Failed.")
		t.FailNow()
	}
}

func makesThings(shouldReturnNil bool) *myTestStruct {
	if !shouldReturnNil {
		return &myTestStruct{}
	}
	return nil
}

func TestShouldBeNil(t *testing.T) {
	assertsToNil := makesThings(true)
	assertsToNotNil := makesThings(false)

	didFail, didFailErrMsg := shouldBeNil(assertsToNil)
	if didFail {
		t.Error(didFailErrMsg)
		t.FailNow()
	}

	didFail, didFailErrMsg = shouldBeNil(assertsToNotNil)
	if !didFail {
		t.Error("shouldBeNil returned did_fail as `true` for a not nil object")
		t.FailNow()
	}
}

func TestShouldNotBeNil(t *testing.T) {
	assertsToNil := makesThings(true)
	assertsToNotNil := makesThings(false)

	didFail, didFailErrMsg := shouldNotBeNil(assertsToNotNil)
	if didFail {
		t.Error(didFailErrMsg)
		t.FailNow()
	}

	didFail, didFailErrMsg = shouldNotBeNil(assertsToNil)
	if !didFail {
		t.Error("shouldNotBeNil returned did_fail as `true` for a not nil object")
		t.FailNow()
	}
}

func TestShouldContain(t *testing.T) {
	shouldNotHaveFailed, _ := shouldContain("is a", "this is a test")
	if shouldNotHaveFailed {
		t.Errorf("shouldConatain failed.")
		t.FailNow()
	}

	shouldHaveFailed, _ := shouldContain("beer", "this is a test")
	if !shouldHaveFailed {
		t.Errorf("shouldConatain failed.")
		t.FailNow()
	}
}

type anyTestObj struct {
	Id   int
	Name string
}

func TestAny(t *testing.T) {
	testObjs := []anyTestObj{{1, "Test"}, {2, "Test2"}, {3, "Foo"}}

	didFail, _ := shouldAny(testObjs, func(obj interface{}) bool {
		if typed, didType := obj.(anyTestObj); didType {
			return strings.HasPrefix(typed.Name, "Foo")
		}
		return false
	})
	if didFail {
		t.Errorf("shouldAny failed.")
		t.FailNow()
	}

	didFail, _ = shouldAny(testObjs, func(obj interface{}) bool {
		if typed, didType := obj.(anyTestObj); didType {
			return strings.HasPrefix(typed.Name, "Bar")
		}
		return false
	})
	if !didFail {
		t.Errorf("shouldAny should have failed.")
		t.FailNow()
	}

	didFail, _ = shouldAny(anyTestObj{1, "test"}, func(obj interface{}) bool {
		return true
	})
	if !didFail {
		t.Errorf("shouldAny should have failed on non-slice target.")
		t.FailNow()
	}
}

func TestAll(t *testing.T) {
	testObjs := []anyTestObj{{1, "Test"}, {2, "Test2"}, {3, "Foo"}}

	didFail, _ := shouldAll(testObjs, func(obj interface{}) bool {
		if typed, didType := obj.(anyTestObj); didType {
			return typed.Id > 0
		}
		return false
	})
	if didFail {
		t.Errorf("shouldAll shouldnt have failed.")
		t.FailNow()
	}

	didFail, _ = shouldAll(testObjs, func(obj interface{}) bool {
		if typed, didType := obj.(anyTestObj); didType {
			return strings.HasPrefix(typed.Name, "Test")
		}
		return false
	})
	if !didFail {
		t.Errorf("shouldAll should have failed.")
		t.FailNow()
	}

	didFail, _ = shouldAll(anyTestObj{1, "test"}, func(obj interface{}) bool {
		return true
	})
	if !didFail {
		t.Errorf("shouldAll should have failed on non-slice target.")
		t.FailNow()
	}
}

func TestNone(t *testing.T) {
	testObjs := []anyTestObj{{1, "Test"}, {2, "Test2"}, {3, "Foo"}}

	didFail, _ := shouldNone(testObjs, func(obj interface{}) bool {
		if typed, didType := obj.(anyTestObj); didType {
			return typed.Id > 4
		}
		return false
	})
	if didFail {
		t.Errorf("shouldAll shouldnt have failed.")
		t.FailNow()
	}

	didFail, _ = shouldNone(testObjs, func(obj interface{}) bool {
		if typed, didType := obj.(anyTestObj); didType {
			return typed.Id > 0
		}
		return false
	})
	if !didFail {
		t.Errorf("shouldNone should have failed.")
		t.FailNow()
	}
}

func TestInTimeDelta(t *testing.T) {
	value1 := time.Date(2016, 1, 29, 9, 0, 0, 0, time.UTC)
	value2 := time.Date(2016, 1, 29, 9, 0, 0, 1, time.UTC)
	value3 := time.Date(2016, 1, 29, 8, 0, 0, 0, time.UTC)
	value4 := time.Date(2015, 1, 29, 9, 0, 0, 0, time.UTC)

	didFail, _ := shouldBeInTimeDelta(value1, value2, 1*time.Minute)
	if didFail {
		t.Errorf("shouldBeInTimeDelta shouldnt have failed.")
		t.FailNow()
	}

	didFail, _ = shouldBeInTimeDelta(value1, value3, 1*time.Minute)
	if !didFail {
		t.Errorf("shouldBeInTimeDelta should have failed.")
		t.FailNow()
	}

	didFail, _ = shouldBeInTimeDelta(value1, value4, 1*time.Minute)
	if !didFail {
		t.Errorf("shouldBeInTimeDelta should have failed.")
		t.FailNow()
	}
}
