package example

import (
	"context"
	"fmt"
	"time"

	"github.com/dop251/goja"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/lib/netext/grpcext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func init() {
	modules.Register("k6/x/chat", new(RootModule))
}

type (
	// RootModule is the global module instance that will create module
	// instances for each VU.
	RootModule struct{}

	// ModuleInstance represents an instance of the module.
	ModuleInstance struct {
		vu modules.VU
	}
)

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &ModuleInstance{}
)

// New returns a pointer to a new RootModule instance.
func New() *RootModule {
	return &RootModule{}
}

// NewModuleInstance implements the modules.Module interface to return
// a new instance for each VU.
func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &ModuleInstance{vu: vu}
}

// Exports returns the exports of the module.
func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"send": mi.OpenAndSend,
		},
	}
}

func (mi *ModuleInstance) OpenAndSend(addr string, method string, payload goja.Value) (*grpcext.Response, error) {
	ctx, cancel := context.WithTimeout(mi.vu.Context(), 5*time.Second)
	defer cancel()

	notls := grpc.WithTransportCredentials(insecure.NewCredentials())

	conn, err := grpcext.Dial(ctx, addr, notls)
	if err != nil {
		return nil, fmt.Errorf("dialing failed: %w", err)
	}
	defer conn.Close()

	md, err := mi.methodDescriptor(ctx, conn, method)
	if err != nil {
		return nil, err
	}

	req := grpcext.Request{
		MethodDescriptor: md,
		Tags:             make(map[string]string),
	}
	if payload != nil {
		b, err := payload.ToObject(mi.vu.Runtime()).MarshalJSON()
		if err != nil {
			return nil, err
		}
		req.Message = b
	}
	return conn.Invoke(ctx, method, nil, req)
}

func (mi *ModuleInstance) methodDescriptor(ctx context.Context, c *grpcext.Conn, method string) (protoreflect.MethodDescriptor, error) {
	rc, err := c.ReflectionClient()
	if err != nil {
		return nil, err
	}

	fdset, err := rc.Reflect(ctx)
	if err != nil {
		return nil, err
	}

	files, err := protodesc.NewFiles(fdset)
	if err != nil {
		return nil, err
	}

	var md protoreflect.MethodDescriptor
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		sds := fd.Services()
		for i := 0; i < sds.Len(); i++ {
			sd := sds.Get(i)
			sdname := sd.FullName()

			mds := sd.Methods()
			for j := 0; j < mds.Len(); j++ {
				mdi := mds.Get(j)
				methodName := fmt.Sprintf("/%s/%s", sdname, mdi.Name())

				if methodName == method {
					md = mdi
					return false
				}
			}
		}
		return true
	})
	if md == nil {
		return nil, fmt.Errorf("service method not found")
	}
	return md, nil
}
