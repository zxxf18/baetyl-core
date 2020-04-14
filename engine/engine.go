package engine

import (
	"errors"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"math/rand"
	"os"
	"time"

	"github.com/baetyl/baetyl-core/ami"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	bh "github.com/timshannon/bolthold"
)

type Engine struct {
	nod  *node.Node
	cfg  config.EngineConfig
	ami  ami.AMI
	tomb utils.Tomb
	log  *log.Logger
	ns   string
}

func NewEngine(cfg config.EngineConfig, sto *bh.Store, nod *node.Node) (*Engine, error) {
	if cfg.Kind != "kubernetes" {
		return nil, os.ErrInvalid
	}
	ami, err := ami.NewKubeImpl(cfg.Kubernetes, sto)
	if err != nil {
		return nil, err
	}
	e := &Engine{
		nod: nod,
		ami: ami,
		cfg: cfg,
		ns:  "baetyl-edge",
		log: log.With(log.Any("engine", cfg.Kind)),
	}
	return e, nil
}

func (e *Engine) Loop() {
	e.tomb.Go(e.reporting)
}

func (e *Engine) Once() error {
	return e.report()
}

func (e *Engine) reporting() error {
	e.log.Info("engine starts to report")
	defer e.log.Info("engine has stopped reporting")

	t := time.NewTicker(e.cfg.Report.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			err := e.report()
			if err != nil {
				e.log.Error("failed to report local shadow", log.Error(err))
			} else {
				e.log.Debug("engine reports local shadow")
			}
		case <-e.tomb.Dying():
			return nil
		}
	}
}

func (e *Engine) report() error {
	// to collect app status
	info, err := e.ami.Collect(e.ns)
	if err != nil {
		return err
	}
	if len(info) == 0 {
		return errors.New("no status collected")
	}
	no, err := e.nod.Get()
	if err != nil {
		return err
	}
	if info["apps"] != nil {
		info["apps"] = alignApps(info.AppInfos(), no.Desire.AppInfos())
	}
	if info["sysapps"] != nil {
		info["sysapps"] = alignApps(info.SysAppInfos(), no.Desire.SysAppInfos())
	}

	// to report app status into local shadow, and return shadow delta
	delta, err := e.nod.Report(info)
	if err != nil {
		return err
	}
	// if apps are updated, to apply new apps
	if delta == nil {
		return nil
	}
	apps := delta.AppInfos()
	if apps == nil {
		return nil
	}
	sysApps := delta.SysAppInfos()
	if sysApps != nil {
		apps = append(apps, sysApps...)
	}
	e.log.Info("to apply apps", log.Any("apps", apps))
	return e.ami.Apply(e.ns, apps)
}

func (e *Engine) Close() {
	e.tomb.Kill(nil)
	e.tomb.Wait()
}

func alignApps(reApps, deApps []v1.AppInfo) []v1.AppInfo {
	if len(reApps) == 0 || len(deApps) == 0 {
		return reApps
	}
	as := map[string]v1.AppInfo{}
	for _, a := range reApps {
		as[a.Name] = a
	}
	var res []v1.AppInfo
	for _, a := range deApps {
		if r, ok := as[a.Name]; ok {
			res = append(res, r)
			delete(as, a.Name)
		}
	}
	for _, a := range as {
		res = append(res, a)
	}
	return res
}
