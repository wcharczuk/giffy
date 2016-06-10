package assert

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	RED    = "31"
	BLUE   = "94"
	GREEN  = "32"
	YELLOW = "33"
	WHITE  = "37"
	GRAY   = "90"

	EMPTY = ""
)

var assertCount int32

func incrementAssertCount() {
	atomic.AddInt32(&assertCount, int32(1))
}

// Count returns the total number of assertions.
func Count() int {
	return int(assertCount)
}

type Predicate func(item interface{}) bool
type PredicateOfInt func(item int) bool
type PredicateOfFloat func(item float64) bool
type PredicateOfString func(item string) bool

type Assertions struct {
	t           *testing.T
	didComplete bool
}

func Empty() *Assertions {
	return &Assertions{}
}

func New(t *testing.T) *Assertions {
	return &Assertions{t: t}
}

func (a *Assertions) assertion() {
	incrementAssertCount()
}

func (a *Assertions) NonFatal() *optional {
	return &optional{a.t}
}

func (a *Assertions) NotNil(object interface{}, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldNotBeNil(object); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) Nil(object interface{}, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldBeNil(object); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) Len(collection interface{}, length int, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldHaveLength(collection, length); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) Empty(collection interface{}, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldBeEmpty(collection); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) NotEmpty(collection interface{}, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldNotBeEmpty(collection); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) Equal(expected interface{}, actual interface{}, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldBeEqual(expected, actual); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) NotEqual(expected interface{}, actual interface{}, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldNotBeEqual(expected, actual); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) Zero(value interface{}, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldBeZero(value); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) NotZero(value interface{}, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldBeNonZero(value); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) True(object bool, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldBeTrue(object); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) False(object bool, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldBeFalse(object); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) InDelta(from, to, delta float64, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldBeInDelta(from, to, delta); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) InTimeDelta(from, to time.Time, delta time.Duration, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldBeInTimeDelta(from, to, delta); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) FileExists(filePath string, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := fileShouldExist(filePath); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) Contains(subString, corpus string, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldContain(subString, corpus); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) Any(target interface{}, predicate Predicate, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldAny(target, predicate); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) AnyOfInt(target []int, predicate PredicateOfInt, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldAnyOfInt(target, predicate); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) AnyOfFloat(target []float64, predicate PredicateOfFloat, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldAnyOfFloat(target, predicate); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) AnyOfString(target []string, predicate PredicateOfString, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldAnyOfString(target, predicate); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) All(target interface{}, predicate Predicate, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldAll(target, predicate); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) AllOfInt(target []int, predicate PredicateOfInt, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldAllOfInt(target, predicate); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) AllOfFloat(target []float64, predicate PredicateOfFloat, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldAllOfFloat(target, predicate); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) AllOfString(target []string, predicate PredicateOfString, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldAllOfString(target, predicate); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) None(target interface{}, predicate Predicate, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldNone(target, predicate); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) NoneOfInt(target []int, predicate PredicateOfInt, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldNoneOfInt(target, predicate); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) NoneOfFloat(target []float64, predicate PredicateOfFloat, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldNoneOfFloat(target, predicate); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) NoneOfString(target []string, predicate PredicateOfString, userMessageComponents ...interface{}) {
	a.assertion()
	if did_fail, message := shouldNoneOfString(target, predicate); did_fail {
		failNow(a.t, message, userMessageComponents...)
	}
}

func (a *Assertions) FailNow(userMessageComponents ...interface{}) {
	failNow(a.t, "Fatal Assertion Failed", userMessageComponents...)
}

func (a *Assertions) StartTimeout(timeout time.Duration, userMessageComponents ...interface{}) {
	sleepFor := 1 * time.Millisecond
	waited := time.Duration(0)
	a.didComplete = false

	go func() {
		for !a.didComplete {
			if waited > timeout {
				panic("Timeout Reached")
			}
			time.Sleep(sleepFor)
			waited += sleepFor
		}
	}()
}

func (a *Assertions) EndTimeout() {
	a.didComplete = true
}

type optional struct {
	t *testing.T
}

func (o *optional) assertion() {
	incrementAssertCount()
}

func (o *optional) Nil(object interface{}, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldBeNil(object); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) NotNil(object interface{}, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldNotBeNil(object); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) Len(collection interface{}, length int, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldHaveLength(collection, length); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) Empty(collection interface{}, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldBeEmpty(collection); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) NotEmpty(collection interface{}, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldNotBeEmpty(collection); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) Equal(expected interface{}, actual interface{}, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldBeEqual(expected, actual); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) NotEqual(expected interface{}, actual interface{}, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldNotBeEqual(expected, actual); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) Zero(value interface{}, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldBeZero(value); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) NotZero(value interface{}, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldBeNonZero(value); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) True(object bool, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldBeTrue(object); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) False(object bool, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldBeFalse(object); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) InDelta(from, to, delta float64, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldBeInDelta(from, to, delta); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) InTimeDelta(from, to time.Time, delta time.Duration, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldBeInTimeDelta(from, to, delta); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) FileExists(filePath string, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := fileShouldExist(filePath); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) Contains(subString, corpus string, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldContain(subString, corpus); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) Any(target interface{}, predicate Predicate, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldAny(target, predicate); did_fail {
		fail(o.t, prefixOptional(message), userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) AnyOfInt(target []int, predicate PredicateOfInt, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldAnyOfInt(target, predicate); did_fail {
		failNow(o.t, message, userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) AnyOfFloat(target []float64, predicate PredicateOfFloat, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldAnyOfFloat(target, predicate); did_fail {
		failNow(o.t, message, userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) AnyOfString(target []string, predicate PredicateOfString, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldAnyOfString(target, predicate); did_fail {
		failNow(o.t, message, userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) All(target interface{}, predicate Predicate, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldAll(target, predicate); did_fail {
		failNow(o.t, message, userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) AllOfInt(target []int, predicate PredicateOfInt, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldAllOfInt(target, predicate); did_fail {
		failNow(o.t, message, userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) AllOfFloat(target []float64, predicate PredicateOfFloat, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldAllOfFloat(target, predicate); did_fail {
		failNow(o.t, message, userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) AllOfString(target []string, predicate PredicateOfString, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldAllOfString(target, predicate); did_fail {
		failNow(o.t, message, userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) None(target interface{}, predicate Predicate, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldNone(target, predicate); did_fail {
		failNow(o.t, message, userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) NoneOfInt(target []int, predicate PredicateOfInt, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldNoneOfInt(target, predicate); did_fail {
		failNow(o.t, message, userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) NoneOfFloat(target []float64, predicate PredicateOfFloat, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldNoneOfFloat(target, predicate); did_fail {
		failNow(o.t, message, userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) NoneOfString(target []string, predicate PredicateOfString, userMessageComponents ...interface{}) bool {
	o.assertion()
	if did_fail, message := shouldNoneOfString(target, predicate); did_fail {
		failNow(o.t, message, userMessageComponents...)
		return false
	}
	return true
}

func (o *optional) Fail(userMessageComponents ...interface{}) {
	fail(o.t, prefixOptional("Assertion Failed"), userMessageComponents...)
}

// --------------------------------------------------------------------------------
// OUTPUT
// --------------------------------------------------------------------------------

func failNow(t *testing.T, message string, userMessageComponents ...interface{}) {
	fail(t, message, userMessageComponents...)
	if t != nil {
		t.FailNow()
	} else {
		os.Exit(1)
	}
}

func fail(t *testing.T, message string, userMessageComponents ...interface{}) {
	error_trace := strings.Join(callerInfo(), "\n\t")

	if len(error_trace) == 0 {
		error_trace = "Unknown"
	}

	assertion_failed_label := color("Assertion Failed!", RED)
	location_label := color("Assert Location", GRAY)
	assertion_label := color("Assertion", GRAY)
	message_label := color("Message", GRAY)

	erasure := fmt.Sprintf("\r%s", getClearString())

	if len(userMessageComponents) != 0 {
		user_message := fmt.Sprint(userMessageComponents...)

		error_format := `%s
%s
%s:
	%s
%s: 
	%s
%s: 
	%s
`
		if t != nil {
			t.Errorf(error_format, erasure, assertion_failed_label, location_label, error_trace, assertion_label, message, message_label, user_message)
		} else {
			fmt.Errorf(error_format, "", assertion_failed_label, location_label, error_trace, assertion_label, message, message_label, user_message)
		}

	} else {
		error_format := `%s
%s
%s: 
	%s
%s: 
	%s
`
		if t != nil {
			t.Errorf(error_format, erasure, assertion_failed_label, location_label, error_trace, assertion_label, message)
		} else {
			fmt.Errorf(error_format, "", assertion_failed_label, location_label, error_trace, assertion_label, message)
		}
	}
}

// --------------------------------------------------------------------------------
// ASSERTION LOGIC
// --------------------------------------------------------------------------------

func shouldHaveLength(collection interface{}, length int) (bool, string) {
	if l := getLength(collection); l != length {
		message := shouldBeMultipleMessage(length, l, "Collection should have length")
		return true, message
	}
	return false, EMPTY
}

func shouldNotBeEmpty(collection interface{}) (bool, string) {
	if l := getLength(collection); l == 0 {
		message := "Should not be empty"
		return true, message
	}
	return false, EMPTY
}

func shouldBeEmpty(collection interface{}) (bool, string) {
	if l := getLength(collection); l != 0 {
		message := shouldBeMessage(collection, "Should be empty")
		return true, message
	}
	return false, EMPTY
}

func shouldBeEqual(expected, actual interface{}) (bool, string) {
	if !areEqual(expected, actual) {
		return true, equalMessage(actual, expected)
	}
	return false, EMPTY
}

func shouldNotBeEqual(expected, actual interface{}) (bool, string) {
	if areEqual(expected, actual) {
		return true, notEqualMessage(actual, expected)
	}
	return false, EMPTY
}

func shouldNotBeNil(object interface{}) (bool, string) {
	if isNil(object) {
		return true, "Should not be nil"
	}
	return false, EMPTY
}

func shouldBeNil(object interface{}) (bool, string) {
	if !isNil(object) {
		return true, shouldBeMessage(object, "Should be nil")
	}
	return false, EMPTY
}

func shouldBeTrue(value bool) (bool, string) {
	if !value {
		return true, "Should be true"
	}
	return false, EMPTY
}

func shouldBeFalse(value bool) (bool, string) {
	if value {
		return true, "Should be false"
	}
	return false, EMPTY
}

func shouldBeZero(value interface{}) (bool, string) {
	if !isZero(value) {
		return true, shouldBeMessage(value, "Should be zero")
	}
	return false, EMPTY
}

func shouldBeNonZero(value interface{}) (bool, string) {
	if isZero(value) {
		return true, "Should be non-zero"
	}
	return false, EMPTY
}

func fileShouldExist(filePath string) (bool, string) {
	_, file_err := os.Stat(filePath)
	if file_err != nil {
		pwd, _ := os.Getwd()
		message := fmt.Sprintf("File doesnt exist: %d, `pwd`: %s", filePath, pwd)
		return true, message
	}
	return false, EMPTY
}

func shouldBeInDelta(from, to, delta float64) (bool, string) {
	diff := math.Abs(from - to)
	if diff > delta {
		message := fmt.Sprintf("Difference of %d and %d should be less than %d", from, to, delta)
		return true, message
	}
	return false, EMPTY
}

func shouldBeInTimeDelta(from, to time.Time, delta time.Duration) (bool, string) {
	var diff time.Duration
	if from.After(to) {
		diff = from.Sub(to)
	} else {
		diff = to.Sub(from)
	}
	if diff > delta {
		message := fmt.Sprintf("Delta of %s and %s should be less than %v", from.Format(time.RFC3339), to.Format(time.RFC3339), delta)
		return true, message
	}
	return false, EMPTY
}

func shouldContain(subString, corpus string) (bool, string) {
	if !strings.Contains(corpus, subString) {
		message := fmt.Sprintf("`%s` should contain `%s`", corpus, subString)
		return true, message
	}
	return false, EMPTY
}

func shouldAny(target interface{}, predicate Predicate) (bool, string) {
	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() != reflect.Slice {
		return true, "`target` is not a slice"
	}

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface()
		if predicate(obj) {
			return false, EMPTY
		}
	}
	return true, "Predicate did not fire for any element in target"
}

func shouldAnyOfInt(target []int, predicate PredicateOfInt) (bool, string) {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(int)
		if predicate(obj) {
			return false, EMPTY
		}
	}
	return true, "Predicate did not fire for any element in target"
}

func shouldAnyOfFloat(target []float64, predicate PredicateOfFloat) (bool, string) {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(float64)
		if predicate(obj) {
			return false, EMPTY
		}
	}
	return true, "Predicate did not fire for any element in target"
}

func shouldAnyOfString(target []string, predicate PredicateOfString) (bool, string) {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(string)
		if predicate(obj) {
			return false, EMPTY
		}
	}
	return true, "Predicate did not fire for any element in target"
}

func shouldAll(target interface{}, predicate Predicate) (bool, string) {
	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() != reflect.Slice {
		return true, "`target` is not a slice"
	}

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface()
		if !predicate(obj) {
			return true, fmt.Sprintf("Predicate failed for element in target: %#v", obj)
		}
	}
	return false, EMPTY
}

func shouldAllOfInt(target []int, predicate PredicateOfInt) (bool, string) {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(int)
		if !predicate(obj) {
			return true, fmt.Sprintf("Predicate failed for element in target: %#v", obj)
		}
	}
	return false, EMPTY
}

func shouldAllOfFloat(target []float64, predicate PredicateOfFloat) (bool, string) {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(float64)
		if !predicate(obj) {
			return true, fmt.Sprintf("Predicate failed for element in target: %#v", obj)
		}
	}
	return false, EMPTY
}

func shouldAllOfString(target []string, predicate PredicateOfString) (bool, string) {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(string)
		if !predicate(obj) {
			return true, fmt.Sprintf("Predicate failed for element in target: %#v", obj)
		}
	}
	return false, EMPTY
}

func shouldNone(target interface{}, predicate Predicate) (bool, string) {
	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() != reflect.Slice {
		return true, "`target` is not a slice"
	}

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface()
		if predicate(obj) {
			return true, fmt.Sprintf("Predicate passed for element in target: %#v", obj)
		}
	}
	return false, EMPTY
}

func shouldNoneOfInt(target []int, predicate PredicateOfInt) (bool, string) {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(int)
		if predicate(obj) {
			return true, fmt.Sprintf("Predicate passed for element in target: %#v", obj)
		}
	}
	return false, EMPTY
}

func shouldNoneOfFloat(target []float64, predicate PredicateOfFloat) (bool, string) {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(float64)
		if predicate(obj) {
			return true, fmt.Sprintf("Predicate passed for element in target: %#v", obj)
		}
	}
	return false, EMPTY
}

func shouldNoneOfString(target []string, predicate PredicateOfString) (bool, string) {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(string)
		if predicate(obj) {
			return true, fmt.Sprintf("Predicate passed for element in target: %#v", obj)
		}
	}
	return false, EMPTY
}

// --------------------------------------------------------------------------------
// UTILITY
// --------------------------------------------------------------------------------

func prefixOptional(message string) string {
	return "(Non-Fatal) " + message
}

func shouldBeMultipleMessage(expected, actual interface{}, message string) string {
	expected_label := color("Expected", WHITE)
	actual_label := color("Actual", WHITE)

	return fmt.Sprintf(`%s
	%s: 	%v
	%s: 	%v`, message, expected_label, expected, actual_label, actual)
}

func shouldBeMessage(object interface{}, message string) string {
	actual_label := color("Actual", WHITE)
	return fmt.Sprintf(`%s
	%s: 	%v`, message, actual_label, object)
}

func notEqualMessage(actual, expected interface{}) string {
	return shouldBeMultipleMessage(expected, actual, "Objects should not be equal")
}

func equalMessage(actual, expected interface{}) string {
	return shouldBeMultipleMessage(expected, actual, "Objects should be equal")
}

func getLength(object interface{}) int {
	if object == nil {
		return 0
	} else if object == "" {
		return 0
	}

	objValue := reflect.ValueOf(object)

	switch objValue.Kind() {
	case reflect.Map:
		fallthrough
	case reflect.Slice, reflect.Chan, reflect.String:
		{
			return objValue.Len()
		}
	}
	return 0
}

func isNil(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}
	return false
}

func isZero(value interface{}) bool {
	return areEqual(0, value)
}

func areEqual(expected, actual interface{}) bool {
	if expected == nil && actual == nil {
		return true
	}
	if (expected == nil && actual != nil) || (expected != nil && actual == nil) {
		return false
	}

	actualType := reflect.TypeOf(actual)
	if actualType == nil {
		return false
	}
	expectedValue := reflect.ValueOf(expected)
	if expectedValue.IsValid() && expectedValue.Type().ConvertibleTo(actualType) {
		return reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual)
	}

	return reflect.DeepEqual(expected, actual)
}

func callerInfo() []string {
	pc := uintptr(0)
	file := ""
	line := 0
	ok := false
	name := ""

	callers := []string{}
	for i := 0; ; i++ {
		pc, file, line, ok = runtime.Caller(i)
		if !ok {
			return nil
		}

		if file == "<autogenerated>" {
			break
		}

		parts := strings.Split(file, "/")
		dir := parts[len(parts)-2]
		file = parts[len(parts)-1]
		if dir != "assert" && dir != "go-assert" && dir != "mock" && dir != "require" {
			callers = append(callers, fmt.Sprintf("%s:%d", file, line))
		}

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}
		name = f.Name()

		// Drop the package
		segments := strings.Split(name, ".")
		name = segments[len(segments)-1]
		if isTest(name, "Test") ||
			isTest(name, "Benchmark") ||
			isTest(name, "Example") {
			break
		}
	}

	return callers
}

func color(input string, colorCode string) string {
	return fmt.Sprintf("\033[%s;01m%s\033[0m", colorCode, input)
}

func reflectTypeName(object interface{}) string {
	return reflect.TypeOf(object).Name()
}

func isTest(name, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	if len(name) == len(prefix) { // "Test" is ok
		return true
	}
	rune, _ := utf8.DecodeRuneInString(name[len(prefix):])
	return !unicode.IsLower(rune)
}

func getClearString() string {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]

	return strings.Repeat(" ", len(fmt.Sprintf("%s:%d:      ", file, line))+2)
}
