package overview

import (
	"testing"

	"github.com/coroot/coroot/model"
	"github.com/stretchr/testify/assert"
)

func TestAppFilter(t *testing.T) {
	w := &model.World{Applications: map[model.ApplicationId]*model.Application{}}
	add := func(cluster, ns, name string, kind model.ApplicationKind, category model.ApplicationCategory) *model.Application {
		app := model.NewApplication(model.NewApplicationId(cluster, ns, kind, name))
		app.Category = category
		app.GetOrCreateInstance(name+"-5d9f7c6b4-x7k2p", nil).GetOrCreateContainer("/k8s/"+ns+"/"+name+"-5d9f7c6b4-x7k2p/"+name, name)
		w.Applications[app.Id] = app
		return app
	}
	add("c1", "shop", "cart", model.ApplicationKindDeployment, model.ApplicationCategoryApplication)
	add("c2", "shop", "cart", model.ApplicationKindDeployment, model.ApplicationCategoryApplication)
	checkout := add("c1", "shop", "checkout", model.ApplicationKindDeployment, model.ApplicationCategoryApplication)
	checkout.Settings = &model.ApplicationSettings{
		Tracing: &model.ApplicationSettingsTracing{Service: "checkout-traces"},
		Logs:    &model.ApplicationSettingsLogs{Service: "checkout-logs"},
	}
	add("c1", "ns1", "frontend", model.ApplicationKindDeployment, model.ApplicationCategoryApplication)
	add("c1", "ns2", "frontend", model.ApplicationKindDeployment, model.ApplicationCategoryApplication)
	add("c1", "monitoring", "prometheus", model.ApplicationKindDeployment, model.ApplicationCategoryMonitoring)
	add("c1", "", "standalone", model.ApplicationKindUnknown, model.ApplicationCategoryMonitoring)
	otel := []string{"cart", "checkout", "frontend", "/k8s/shop/cart"}
	shop := &AppFilter{Categories: []model.ApplicationCategory{model.ApplicationCategoryMonitoring}, Namespaces: []string{"shop"}}
	monitoring := &AppFilter{Categories: []model.ApplicationCategory{model.ApplicationCategoryMonitoring}}

	assert.False(t, (*AppFilter)(nil).active())
	assert.False(t, (&AppFilter{}).active())
	assert.Equal(t, []string{"cart", "checkout-traces"}, shop.traceServices(w, "c1", otel))
	assert.Equal(t, []string{"cart"}, shop.traceServices(w, "c2", otel))
	assert.Empty(t, shop.traceServices(w, "c3", otel))
	assert.Empty(t, (&AppFilter{Namespaces: []string{"ns1"}}).traceServices(w, "c1", otel))
	assert.Equal(t, []string{"/k8s/monitoring/prometheus"}, monitoring.logServices(w, "c1", nil, true, false))
	assert.Equal(t, []string{"/k8s/shop/cart", "/k8s/shop/checkout", "cart", "checkout-logs"}, shop.logServices(w, "c1", otel, true, true))
	assert.Equal(t, []string{"/k8s/shop/cart", "/k8s/shop/checkout"}, shop.logServices(w, "c1", otel, true, false))
	assert.Equal(t, []string{"cart"}, shop.logServices(w, "c2", otel, false, true))
	assert.Len(t, applicationRefs(w), 6)
}
