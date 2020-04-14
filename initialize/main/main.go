package main

import (
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/engine"
	"github.com/baetyl/baetyl-core/initialize"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-core/sync"
	"github.com/baetyl/baetyl-go/context"
	bh "github.com/timshannon/bolthold"
	"os"
)

type core struct {
	cfg config.Config
	sto *bh.Store
	sha *node.Node
	eng *engine.Engine
	syn *sync.Sync
}

// NewCore creats a new core
func NewCore(ctx context.Context) (*core, error) {
	var cfg config.Config
	err := ctx.LoadCustomConfig(&cfg)
	if err != nil {
		return nil, err
	}

	c := &core{}
	c.sto, err = store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(cfg.Sync.Cloud.HTTP.Cert); os.IsNotExist(err) {
		i, err := initialize.NewInit(&cfg, c.sto)
		if err != nil {
			i.Close()
			return nil, err
		}
		i.Start()
		i.WaitAndClose()
	}

	c.sha, err = node.NewNode(c.sto)
	if err != nil {
		c.Close()
		return nil, err
	}
	c.eng, err = engine.NewEngine(cfg.Engine, c.sto, c.sha)
	if err != nil {
		return nil, err
	}
	c.syn, err = sync.NewSync(cfg.Sync, c.sto, c.sha)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *core) Close() {
	if c.sto != nil {
		c.sto.Close()
	}
}

func main() {
	context.Run(func(ctx context.Context) error {
		c, err := NewCore(ctx)
		if err != nil {
			return err
		}
		defer c.Close()

		err = c.syn.ReportAndDesire()
		if err != nil {
			return err
		}
		err = c.eng.ReportAndDesire()
		if err != nil {
			return err
		}

		return nil
	})
}
