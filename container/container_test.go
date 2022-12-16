package container

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type AnySliceType struct {
	anyField int
}

func (a *AnySliceType) String() string {
	return ""
}

func AnySliceConstructFunction(t []*DependencyType) *AnySliceType {
	return &AnySliceType{}
}

func TestContainer_Start(t *testing.T) {
	c := New()

	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0), Scoped(PrototypeScope))
	assert.Nil(t, err)
	assert.NotNil(t, def)
	c.DefinitionRegistry().Add(def)

	err = c.Start()
	assert.Nil(t, err)
}

func TestContainer_DefinitionRegistry(t *testing.T) {
	c := New()
	definitionRegistry := c.DefinitionRegistry()
	assert.NotNil(t, definitionRegistry)
}

func TestContainer_InstanceRegistry(t *testing.T) {
	c := New()
	instanceRegistry := c.InstanceRegistry()
	assert.NotNil(t, instanceRegistry)
}

func TestContainer_Hooks(t *testing.T) {
	c := New()
	hooks := c.Hooks()
	assert.NotNil(t, hooks)
}

func TestContainer_Get(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0))
	assert.Nil(t, err)
	assert.NotNil(t, def)

	c.DefinitionRegistry().Add(def)

	var instance any
	instance, err = c.Get(context.Background(), "anyType")
	assert.Nil(t, err)
	assert.NotNil(t, instance)
}

func TestContainer_GetInstanceWithSliceDependency(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnySliceConstructFunction)
	assert.Nil(t, err)
	assert.NotNil(t, def)

	c.DefinitionRegistry().Add(def)
	c.InstanceRegistry().Add("test1", &DependencyType{})
	c.InstanceRegistry().Add("test2", &DependencyType{})

	var instance any
	instance, err = c.Get(context.Background(), "anySliceType")
	assert.Nil(t, err)
	assert.NotNil(t, instance)
}

func TestContainer_GetByType(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0))
	assert.Nil(t, err)
	assert.NotNil(t, def)

	c.DefinitionRegistry().Add(def)

	var instance any
	instance, err = c.GetByType(context.Background(), TypeOf[*AnyType]())
	assert.Nil(t, err)
	assert.NotNil(t, instance)
}

func TestContainer_GetByNameAndTypeReturnsInstance(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0))
	assert.Nil(t, err)
	assert.NotNil(t, def)

	c.DefinitionRegistry().Add(def)

	var instance any
	instance, err = c.GetByNameAndType(context.Background(), "anyType", TypeOf[*AnyType]())
	assert.Nil(t, err)
	assert.NotNil(t, instance)
}

func TestContainer_GetByNameAndTypeReturnsErrorIfDefinitionTypeWithGivenNameDoesNotMatchRequiredType(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0))
	assert.Nil(t, err)
	assert.NotNil(t, def)

	c.DefinitionRegistry().Add(def)

	def, err = MakeDefinition(AnySliceConstructFunction)
	assert.Nil(t, err)
	assert.NotNil(t, def)

	c.DefinitionRegistry().Add(def)

	var instance any
	instance, err = c.GetByNameAndType(context.Background(), "anySliceType", TypeOf[*AnyType]())
	assert.NotNil(t, err)
	assert.Nil(t, instance)
	assert.Equal(t, "container: definition type with name anySliceType does not match the required type", err.Error())
}

func TestContainer_GetByNameAndArgsReturnsInstanceWithArguments(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0))
	assert.Nil(t, err)
	assert.NotNil(t, def)

	c.DefinitionRegistry().Add(def)

	depType := &DependencyType{}

	var instance any
	instance, err = c.GetByNameAndArgs(context.Background(), "anyType", depType)
	assert.Nil(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, depType, instance.(*AnyType).t)
}

func TestContainer_GetByNameAndArgsReturnsErrorIfNumberOfProvidedArgumentIsWrong(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0))
	assert.Nil(t, err)
	assert.NotNil(t, def)

	c.DefinitionRegistry().Add(def)

	depType := &DependencyType{}

	var instance any
	instance, err = c.GetByNameAndArgs(context.Background(), "anyType", depType, "anyString")
	assert.NotNil(t, err)
	assert.Nil(t, instance)
	assert.Equal(t, "container: the number of provided arguments is wrong for definition anyType", err.Error())
}

func TestContainer_GetInstancesReturnsInstancesForRequiredTypes(t *testing.T) {
	c := New()
	anyInstance := &AnyType{}
	c.InstanceRegistry().Add("instance", anyInstance)
	anotherInstance := &AnyType{}
	c.InstanceRegistry().Add("anotherInstance", anotherInstance)

	instances, err := c.GetInstancesByType(context.Background(), TypeOf[*AnyType]())
	assert.Nil(t, err)
	assert.NotNil(t, instances)
	assert.Equal(t, []any{anyInstance, anotherInstance}, instances)
}

func TestContainer_GetInstancesByTypeReturnsErrorIfRequiredTypeIsNil(t *testing.T) {
	c := New()
	instance, err := c.GetInstancesByType(context.Background(), nil)
	assert.NotNil(t, err)
	assert.Nil(t, instance)
	assert.Equal(t, "container: requiredType cannot be nil", err.Error())
}

func TestContainer_IsPrototype(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0), Scoped(PrototypeScope))
	assert.Nil(t, err)
	assert.NotNil(t, def)
	c.DefinitionRegistry().Add(def)

	assert.True(t, c.IsPrototype("anyType"))
}

func TestContainer_IsShared(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0))
	assert.Nil(t, err)
	assert.NotNil(t, def)
	c.DefinitionRegistry().Add(def)

	assert.True(t, c.IsShared("anyType"))
}

func TestContainer_PostConstructorShouldBeCalled(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0))
	assert.Nil(t, err)
	assert.NotNil(t, def)

	c.DefinitionRegistry().Add(def)

	var instance any
	instance, err = c.Get(context.Background(), "anyType")

	assert.Nil(t, err)
	assert.NotNil(t, instance)
	instance.(*AnyType).AssertExpectations(t)
}

func TestContainer_ContainsChecksIfInstanceExists(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0))
	assert.Nil(t, err)
	assert.NotNil(t, def)

	c.DefinitionRegistry().Add(def)

	assert.False(t, c.Contains("anyType"))

	var instance any
	instance, err = c.Get(context.Background(), "anyType")

	assert.Nil(t, err)
	assert.NotNil(t, instance)

	assert.True(t, c.Contains("anyType"))
}

func TestContainer_GetReturnsDifferentInstanceForEachCallIfTypeScopeIsPrototype(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0), Scoped(PrototypeScope))
	assert.Nil(t, err)
	assert.NotNil(t, def)
	c.DefinitionRegistry().Add(def)

	var instance any
	instance, err = c.Get(context.Background(), "anyType")
	assert.Nil(t, err)
	assert.NotNil(t, instance)

	var anotherInstance any
	anotherInstance, err = c.Get(context.Background(), "anyType")
	assert.Nil(t, err)
	assert.NotNil(t, anotherInstance)

	assert.NotEqual(t, instance, anotherInstance)
}

func TestContainer_HooksShouldBeCalledDuringInitialization(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0), Scoped(PrototypeScope))
	assert.Nil(t, err)
	assert.NotNil(t, def)
	c.DefinitionRegistry().Add(def)

	preHookCalled := false
	err = c.Hooks().Add(PreInitialization(func(name string, instance any) (any, error) {
		preHookCalled = true
		return instance, nil
	}))
	assert.Nil(t, err)

	postHookCalled := false
	err = c.Hooks().Add(PostInitialization(func(name string, instance any) (any, error) {
		postHookCalled = true
		return instance, nil
	}))
	assert.Nil(t, err)

	var instance any
	instance, err = c.Get(context.Background(), "anyType")
	assert.Nil(t, err)
	assert.NotNil(t, instance)

	assert.True(t, preHookCalled)
	assert.True(t, postHookCalled)
}

func TestContainer_GetByTypeReturnsErrorIfThereIsMoreThanOneDefinition(t *testing.T) {
	c := New()
	def, err := MakeDefinition(AnyConstructFunction, OptionalAt(0), Scoped(PrototypeScope))
	assert.Nil(t, err)
	assert.NotNil(t, def)
	c.DefinitionRegistry().Add(def)

	def, err = MakeDefinition(AnySliceConstructFunction, OptionalAt(0), Scoped(PrototypeScope))
	assert.Nil(t, err)
	assert.NotNil(t, def)
	c.DefinitionRegistry().Add(def)

	var instance any
	instance, err = c.GetByType(context.Background(), TypeOf[fmt.Stringer]())

	assert.NotNil(t, err)
	assert.Equal(t, "container: there is more than one definition for the required type Stringer, it cannot be distinguished", err.Error())
	assert.Nil(t, instance)
}

func TestContainer_GetReturnsErrorIfDefinitionDoesNotExist(t *testing.T) {
	c := New()
	instance, err := c.Get(context.Background(), "any")

	assert.NotNil(t, err)
	assert.Equal(t, "container: not found definition with name any", err.Error())
	assert.Nil(t, instance)
}