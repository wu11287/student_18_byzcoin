package onet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"

	bolt "github.com/coreos/bbolt"
	"gopkg.in/dedis/onet.v2/log"
	"gopkg.in/dedis/onet.v2/network"
)

// Context represents the methods that are available to a service.
type Context struct {
	overlay           *Overlay
	server            *Server
	serviceID         ServiceID
	manager           *serviceManager
	bucketName        []byte
	bucketVersionName []byte
}

// defaultContext is the implementation of the Context interface. It is
// instantiated for each Service.
func newContext(c *Server, o *Overlay, servID ServiceID, manager *serviceManager) *Context {
	ctx := &Context{
		overlay:           o,
		server:            c,
		serviceID:         servID,
		manager:           manager,
		bucketName:        []byte(ServiceFactory.Name(servID)),
		bucketVersionName: []byte(ServiceFactory.Name(servID) + "version"),
	}
	err := manager.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(ctx.bucketName)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(ctx.bucketVersionName)
		return err
	})
	if err != nil {
		log.Panic("Failed to create bucket: " + err.Error())
	}
	return ctx
}

// NewTreeNodeInstance creates a TreeNodeInstance that is bound to a
// service instead of the Overlay.
func (c *Context) NewTreeNodeInstance(t *Tree, tn *TreeNode, protoName string) *TreeNodeInstance {
	io := c.overlay.protoIO.getByName(protoName)
	return c.overlay.NewTreeNodeInstanceFromService(t, tn, ProtocolNameToID(protoName), c.serviceID, io)
}

// SendRaw sends a message to the ServerIdentity.
func (c *Context) SendRaw(si *network.ServerIdentity, msg interface{}) error {
	_, err := c.server.Send(si, msg)
	return err
}

// ServerIdentity returns this server's identity.
func (c *Context) ServerIdentity() *network.ServerIdentity {
	return c.server.ServerIdentity
}

// Suite returns the suite for the context's associated server.
func (c *Context) Suite() network.Suite {
	return c.server.Suite()
}

// ServiceID returns the service-id.
func (c *Context) ServiceID() ServiceID {
	return c.serviceID
}

// CreateProtocol returns a ProtocolInstance bound to the service.
func (c *Context) CreateProtocol(name string, t *Tree) (ProtocolInstance, error) {
	pi, err := c.overlay.CreateProtocol(name, t, c.serviceID)
	return pi, err
}

// ProtocolRegister signs up a new protocol to this Server. Contrary go
// GlobalProtocolRegister, the protocol registered here is tied to that server.
// This is useful for simulations where more than one Server exists in the
// global namespace.
// It returns the ID of the protocol.
func (c *Context) ProtocolRegister(name string, protocol NewProtocol) (ProtocolID, error) {
	return c.server.ProtocolRegister(name, protocol)
}

// RegisterProtocolInstance registers a new instance of a protocol using overlay.
func (c *Context) RegisterProtocolInstance(pi ProtocolInstance) error {
	return c.overlay.RegisterProtocolInstance(pi)
}

// ReportStatus returns all status of the services.
func (c *Context) ReportStatus() map[string]*Status {
	return c.server.statusReporterStruct.ReportStatus()
}

// RegisterStatusReporter registers a new StatusReporter.
func (c *Context) RegisterStatusReporter(name string, s StatusReporter) {
	c.server.statusReporterStruct.RegisterStatusReporter(name, s)
}

// RegisterProcessor overrides the RegisterProcessor methods of the Dispatcher.
// It delegates the dispatching to the serviceManager.
func (c *Context) RegisterProcessor(p network.Processor, msgType network.MessageTypeID) {
	c.manager.registerProcessor(p, msgType)
}

// RegisterProcessorFunc takes a message-type and a function that will be called
// if this message-type is received.
func (c *Context) RegisterProcessorFunc(msgType network.MessageTypeID, fn func(*network.Envelope)) {
	c.manager.registerProcessorFunc(msgType, fn)
}

// RegisterMessageProxy registers a message proxy only for this server /
// overlay
func (c *Context) RegisterMessageProxy(m MessageProxy) {
	c.overlay.RegisterMessageProxy(m)
}

// Service returns the corresponding service.
func (c *Context) Service(name string) Service {
	return c.manager.service(name)
}

// String returns the host it's running on.
func (c *Context) String() string {
	return c.server.ServerIdentity.String()
}

var testContextData = struct {
	service map[string][]byte
	sync.Mutex
}{service: make(map[string][]byte, 0)}

// The ContextDB interface allows for easy testing in the services.
type ContextDB interface {
	Load(key []byte) (interface{}, error)
	LoadRaw(key []byte) ([]byte, error)
	LoadVersion() (int, error)
	SaveVersion(version int) error
}

// Save takes a key and an interface. The interface will be network.Marshal'ed
// and saved in the database under the bucket named after the service name.
//
// The data will be stored in a different bucket for every service.
func (c *Context) Save(key []byte, data interface{}) error {
	buf, err := network.Marshal(data)
	if err != nil {
		return err
	}
	return c.manager.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.bucketName)
		return b.Put(key, buf)
	})
}

// Load takes a key and returns the network.Unmarshaled data.
// Returns a nil value if the key does not exist.
func (c *Context) Load(key []byte) (interface{}, error) {
	var buf []byte
	err := c.manager.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(c.bucketName).Get(key)
		if v == nil {
			return nil
		}

		buf = make([]byte, len(v))
		copy(buf, v)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if buf == nil {
		return nil, nil
	}

	_, ret, err := network.Unmarshal(buf, c.server.suite)
	return ret, err
}

// LoadRaw takes a key and returns the raw, unmarshalled data.
// Returns a nil value if the key does not exist.
func (c *Context) LoadRaw(key []byte) ([]byte, error) {
	var buf []byte
	err := c.manager.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(c.bucketName).Get(key)
		if v == nil {
			return nil
		}

		buf = make([]byte, len(v))
		copy(buf, v)
		return nil
	})
	return buf, err
}

var dbVersion = []byte("dbVersion")

// LoadVersion returns the version of the database, or 0 if
// no version has been found.
func (c *Context) LoadVersion() (int, error) {
	var buf []byte
	err := c.manager.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(c.bucketVersionName).Get(dbVersion)
		if v == nil {
			return nil
		}

		buf = make([]byte, len(v))
		copy(buf, v)
		return nil
	})

	if err != nil {
		return -1, err
	}

	if len(buf) == 0 {
		return 0, nil
	}
	var version int32
	err = binary.Read(bytes.NewReader(buf), binary.LittleEndian, &version)
	return int(version), err
}

// SaveVersion stores the given version as the current database version.
func (c *Context) SaveVersion(version int) error {
	buf := bytes.NewBuffer(nil)
	err := binary.Write(buf, binary.LittleEndian, int32(version))
	if err != nil {
		return err
	}
	return c.manager.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.bucketVersionName)
		return b.Put(dbVersion, buf.Bytes())
	})
}

// GetAdditionalBucket makes sure that a bucket with the given name
// exists, by eventually creating it, and returns the created bucket name,
// which is the servicename + "_" + the given name.
//
// This function should only be used if the Load and Save functions are not sufficient.
// Additionally, the user should not create buckets directly on the DB but always
// call this function to create new buckets to avoid bucket name conflicts.
func (c *Context) GetAdditionalBucket(name []byte) (*bolt.DB, []byte) {
	fullName := append(append(c.bucketName, byte('_')), name...)
	err := c.manager.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(fullName)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return c.manager.db, fullName
}
