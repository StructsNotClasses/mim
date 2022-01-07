package instance

import (
	"github.com/StructsNotClasses/musicplayer/instance/dirtree"
	"github.com/StructsNotClasses/musicplayer/instance/notification"
	"github.com/StructsNotClasses/musicplayer/instance/playback"
	"github.com/StructsNotClasses/musicplayer/musicarray"
	"github.com/StructsNotClasses/musicplayer/remote"

	gnc "github.com/rthornton128/goncurses"

	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

const INT32MAX = 2147483647

type Script []byte

type InputMode int

const (
	CommandMode InputMode = iota
	CharacterMode
)

type UserStateMachine struct {
	line               []byte
	lines              []byte
	bindChar           gnc.Key
	onPlaybackBeingSet bool
	mode               InputMode
	exit               bool
}

type Instance struct {
	client          Client
	tree            dirtree.DirTree
	currentRemote   remote.Remote
	bindMap         map[gnc.Key]Script
	commandMap      map[string]string
	runOnNoPlayback []byte
	state           UserStateMachine
	playbackState   playback.PlaybackState
	notifier        chan notification.Notification
}

func New(scr *gnc.Window, array musicarray.MusicArray) Instance {
	//make user input non-blocking
	scr.Timeout(0)

	if len(array) > INT32MAX {
		log.Fatal(fmt.Sprintf("Cannot play more than %d songs at a time.", INT32MAX))
	}

	rand.Seed(time.Now().UnixNano())

	var instance Instance
	client, err := createClient(scr)
	if err != nil {
		log.Fatal(err)
	}
	instance.client = client
	instance.tree = dirtree.New(array)
	instance.currentRemote = remote.Remote{}

	instance.bindMap = make(map[gnc.Key]Script)
	instance.commandMap = make(map[string]string)
	instance.notifier = make(chan notification.Notification)
	instance.state = UserStateMachine{
		bindChar:           0,
		onPlaybackBeingSet: false,
		mode:               CommandMode,
		exit:               false,
	}

	return instance
}

func (instance *Instance) GetKey() gnc.Key {
	return instance.client.backgroundWindow.GetChar()
}

func (instance *Instance) LoadConfig(filename string) error {
	instance.state.line = []byte{}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	for _, b := range bytes {
		instance.HandleKey(gnc.Key(b))
	}
	return nil
}
