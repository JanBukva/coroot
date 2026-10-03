package overview

import (
	"slices"
	"strings"

	"github.com/coroot/coroot/model"
	"github.com/coroot/coroot/utils"
)

type AppFilter struct {
	Categories []model.ApplicationCategory `json:"categories"`
	Namespaces []string                    `json:"namespaces"`
}

func (f *AppFilter) active() bool {
	return f != nil && (len(f.Categories) > 0 || len(f.Namespaces) > 0)
}

func (f *AppFilter) matches(app *model.Application, clusterId string) bool {
	if app.Id.ClusterId != clusterId || !isListed(app) {
		return false
	}
	if len(f.Namespaces) > 0 {
		return slices.Contains(f.Namespaces, app.Id.Namespace)
	}
	return slices.Contains(f.Categories, app.Category)
}

func (f *AppFilter) traceServices(w *model.World, clusterId string, services []string) []string {
	services = slices.DeleteFunc(services, func(s string) bool { return strings.HasPrefix(s, "/") })
	guessed := model.GuessServices(services, w, clusterId)
	res := utils.NewStringSet()
	for _, app := range w.Applications {
		if !f.matches(app, clusterId) {
			continue
		}
		if app.Settings != nil && app.Settings.Tracing != nil {
			res.Add(app.Settings.Tracing.Service)
		} else {
			res.Add(guessed[app.Id])
		}
	}
	return res.Items()
}

func (f *AppFilter) logServices(w *model.World, clusterId string, otelServices []string, agent, otel bool) []string {
	guessed := model.GuessServices(otelServices, w, clusterId)
	res := utils.NewStringSet()
	for _, app := range w.Applications {
		if !f.matches(app, clusterId) {
			continue
		}
		if agent {
			res.Add(app.LogServices()...)
		}
		if otel {
			if app.Settings != nil && app.Settings.Logs != nil {
				res.Add(app.Settings.Logs.Service)
			} else {
				res.Add(guessed[app.Id])
			}
		}
	}
	return res.Items()
}

type ApplicationRef struct {
	Id       model.ApplicationId       `json:"id"`
	Category model.ApplicationCategory `json:"category"`
}

func applicationRefs(w *model.World) []ApplicationRef {
	var res []ApplicationRef
	for _, app := range w.Applications {
		if isListed(app) {
			res = append(res, ApplicationRef{Id: app.Id, Category: app.Category})
		}
	}
	return res
}
