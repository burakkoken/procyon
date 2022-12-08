package container

import (
	"errors"
	"fmt"
	"github.com/procyon-projects/reflector"
)

type Container struct {
	definitionRegistry *DefinitionRegistry
	instanceRegistry   *InstanceRegistry
	hooks              *Hooks
}

func New() *Container {
	return &Container{
		definitionRegistry: NewDefinitionRegistry(copyDefinitions()),
		instanceRegistry:   NewInstanceRegistry(),
		hooks:              NewHooks(),
	}
}

func (c *Container) Start() error {
	return nil
}

func (c *Container) DefinitionRegistry() *DefinitionRegistry {
	return c.definitionRegistry
}

func (c *Container) InstanceRegistry() *InstanceRegistry {
	return c.instanceRegistry
}

func (c *Container) Hooks() *Hooks {
	return c.hooks
}

func (c *Container) Get(name string) (any, error) {
	return c.getInstance(name, nil)
}

func (c *Container) GetByNameAndType(name string, typ *Type) (any, error) {
	return c.getInstance(name, typ)
}

func (c *Container) GetByNameAndArgs(name string, args ...any) (any, error) {
	return c.getInstance(name, nil, args...)
}

func (c *Container) GetByType(typ *Type) (any, error) {
	return c.getInstance("", typ)
}

func (c *Container) GetInstancesByType(requiredType *Type) ([]any, error) {
	if requiredType == nil {
		return nil, errors.New("container: requiredType cannot be nil")
	}

	instances, err := c.getInstances(reflector.ToSlice(reflector.TypeOf[[]any]()), requiredType)

	if err != nil {
		return nil, err
	}

	return instances.([]any), nil
}

func (c *Container) Contains(name string) bool {
	return c.instanceRegistry.Contains(name)
}

func (c *Container) IsShared(name string) bool {
	def, ok := c.definitionRegistry.Find(name)
	return ok && def.IsShared()
}

func (c *Container) IsPrototype(name string) bool {
	def, ok := c.definitionRegistry.Find(name)
	return ok && def.IsPrototype()
}

func (c *Container) getInstance(name string, requiredType *Type, args ...any) (any, error) {
	if name == "" && requiredType == nil {
		return nil, errors.New("container: either name or requiredType should be given")
	}

	if name == "" {
		candidate, err := c.instanceRegistry.FindByType(requiredType)

		if err == nil {
			return candidate, nil
		}

		names := c.definitionRegistry.DefinitionNamesByType(requiredType)

		if len(names) == 0 {
			return nil, &notFoundError{
				ErrorString: fmt.Sprintf("container: not found instance or definition with required type %s", requiredType.Name()),
			}
		} else if len(names) > 1 {
			return nil, fmt.Errorf("container: there is more than one definition for the required type %s, it cannot be distinguished ", requiredType.Name())
		}

		name = names[0]
	}

	def, ok := c.definitionRegistry.Find(name)

	if !ok {
		return nil, &notFoundError{
			ErrorString: fmt.Sprintf("container: not found definition with name %s", name),
		}
	}

	if requiredType != nil && !c.match(def.reflectorType(), requiredType.typ) {
		return nil, fmt.Errorf("container: definition type with name %s does not match the required type", name)
	}

	if def.IsShared() {
		instance, err := c.instanceRegistry.OrElseGet(name, func() (any, error) {
			instance, err := c.createInstance(def, args)

			if err != nil {
				return nil, err
			}

			return instance, c.instanceRegistry.Add(name, instance)
		})

		return instance, err
	} else if def.IsPrototype() {
		return c.createInstance(def, args)
	}

	return nil, nil
}

func (c *Container) match(instanceType reflector.Type, requiredType reflector.Type) bool {
	if instanceType.CanConvert(requiredType) {
		return true
	} else if reflector.IsPointer(instanceType) && !reflector.IsPointer(requiredType) && !reflector.IsInterface(requiredType) {
		ptrType := reflector.ToPointer(instanceType)

		if ptrType.Elem().CanConvert(requiredType) {
			return true
		}
	}

	return false
}

func (c *Container) createInstance(definition *Definition, args []any) (instance any, err error) {
	newFunc := definition.constructorFunc
	parameterCount := len(definition.Inputs())

	if parameterCount != 0 && len(args) == 0 {
		var resolvedArguments []any
		resolvedArguments, err = c.resolveInputs(definition.Inputs())

		if err != nil {
			return nil, err
		}

		var results []any
		results, err = newFunc.Invoke(resolvedArguments...)

		if err != nil {
			return nil, err
		}

		instance = results[0]
	} else if (parameterCount == 0 && len(args) == 0) || (len(args) != 0 && parameterCount == len(args)) {
		var results []any
		results, err = newFunc.Invoke(args...)

		if err != nil {
			return nil, err
		}

		instance = results[0]
	} else {
		return nil, fmt.Errorf("container: the number of provided arguments is wrong for definition %s", definition.Name())
	}

	return c.initializeInstance(definition.name, instance)
}

func (c *Container) getInstances(sliceType reflector.Slice, itemType *Type) (any, error) {
	val, err := sliceType.Instantiate()

	if err != nil {
		return nil, err
	}

	var (
		instance any
		items    any
	)

	instances := c.instanceRegistry.FindAllByType(itemType)

	sliceType = reflector.ToSlice(reflector.ToPointer(reflector.TypeOfAny(val.Val())).Elem())
	items, err = sliceType.Append(instances...)

	if err != nil {
		return nil, err
	}

	definitionNames := c.definitionRegistry.DefinitionNamesByType(itemType)

	for _, definitionName := range definitionNames {
		if c.InstanceRegistry().Contains(definitionName) {
			continue
		}

		instance, err = c.Get(definitionName)

		if err != nil {
			return nil, err
		}

		items, err = sliceType.Append(instance)
	}

	return items, nil
}

func (c *Container) resolveInputs(inputs []*Input) ([]any, error) {
	arguments := make([]any, 0)
	for _, input := range inputs {

		if reflector.IsSlice(input.reflectorType()) {
			sliceType := reflector.ToSlice(input.reflectorType())
			instances, err := c.getInstances(sliceType, &Type{
				typ: sliceType.Elem(),
			})

			if err != nil {
				return nil, err
			}

			arguments = append(arguments, instances)
			continue
		}

		var (
			instance any
			err      error
		)

		if input.Name() != "" {
			instance, err = c.Get(input.Name())
		} else {
			instance, err = c.GetByType(input.Type())
		}

		if err != nil {
			if notFoundErr := (*notFoundError)(nil); errors.As(err, &notFoundErr) && !reflector.IsPointer(input.reflectorType()) && input.reflectorType().IsInstantiable() {
				var val reflector.Value
				val, err = input.reflectorType().Instantiate()

				if err != nil {
					return nil, err
				}

				instance = val.Elem()
				arguments = append(arguments, instance)
				continue
			}

			if !input.IsOptional() && err != nil {
				return nil, err
			} else if input.IsOptional() && err != nil {
				arguments = append(arguments, nil)
			}
		} else {
			arguments = append(arguments, instance)
		}
	}

	return arguments, nil
}

func (c *Container) initializeInstance(name string, instance any) (any, error) {
	var (
		err error
	)
	hooks := c.Hooks().ToSlice()

	for _, hook := range hooks {
		if hook.OnPreInitialization != nil {
			instance, err = hook.OnPreInitialization(name, instance)
		}

		if err != nil {
			return nil, err
		}
	}

	if postConstructor, implements := instance.(PostConstructor); implements {
		err = postConstructor.PostConstruct()

		if err != nil {
			return nil, err
		}
	}

	for _, hook := range hooks {
		if hook.OnPostInitialization != nil {
			instance, err = hook.OnPostInitialization(name, instance)

			if err != nil {
				return nil, err
			}
		}
	}

	return instance, nil
}
