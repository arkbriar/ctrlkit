package controller

import (
	"context"
	"testing"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apiv1 "demo/api/v1"
)

var scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = apiv1.AddToScheme(scheme)
}

func Test_CronJobController_Reconcile(t *testing.T) {
	controller := &CronJobController{
		Client: fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(&apiv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}).
			Build(),
		Logger: zapr.NewLogger(zap.NewExample()),
	}

	_, err := controller.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "example",
			Namespace: "default",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}
