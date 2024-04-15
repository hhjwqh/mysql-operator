/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	mysqlv1 "github.com/hhjwqh/mysql-operator/api/v1"
	"github.com/hhjwqh/mysql-operator/controllers/utils"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

// MysqlrwhaReconciler reconciles a Mysqlrwha object
type MysqlrwhaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mysql.github.com,resources=mysqlrwhas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.github.com,resources=mysqlrwhas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql.github.com,resources=mysqlrwhas/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Mysqlrwha object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *MysqlrwhaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	app := &mysqlv1.Mysqlrwha{}
	//从缓存中获取app
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//根据app的配置进行处理
	//1.mysqlconfigmap处理
	mysqlconfigmap := utils.NewMysqlConfigmap(app)
	if err := controllerutil.SetControllerReference(app, mysqlconfigmap, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	//查找同名mysqlconfigmap
	c := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: app.Name + "-mysql-configmap", Namespace: app.Namespace}, c); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, mysqlconfigmap); err != nil {
				logger.Error(err, "create mysqlconfigmap failed")
				return ctrl.Result{}, err
			}
		}
	} else {
		if err := r.Update(ctx, mysqlconfigmap); err != nil {
			return ctrl.Result{}, err
		}
	}

	//2. headlessservice的处理
	headlessservice := utils.NewHeadlessService(app)
	if err := controllerutil.SetControllerReference(app, headlessservice, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	//查找同名headlessservice
	hl := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: app.Name + "-headless", Namespace: app.Namespace}, hl); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, headlessservice); err != nil {
				logger.Error(err, "create headlessservice failed")
				return ctrl.Result{}, err
			}
		}
	} else {
		if err := r.Update(ctx, headlessservice); err != nil {
			return ctrl.Result{}, err
		}
	}
	//3. mysqlstatefulset的处理
	statefulset := utils.NewStatefullset(app)
	for _, v := range statefulset.Spec.Template.Spec.Containers {
		for i, _ := range v.Command {
			v.Command[i] = strings.Replace(v.Command[i], "Mysql-Master-headless", app.Name+"-mysql-0."+app.Name+"-headless", -1)
			v.Command[i] = strings.Replace(v.Command[i], "MysqlRootPassword", app.Spec.Mysql.MysqlRootPassword, -1)
			v.Command[i] = strings.Replace(v.Command[i], "ObjectMeta-Pod-Name", app.Name+"-mysql", -1)
			v.Command[i] = strings.Replace(v.Command[i], "ObjectMeta-Name-headless", app.Name+"-headless", -1)
		}
		if v.LivenessProbe != nil && v.LivenessProbe.Exec != nil && len(v.LivenessProbe.Exec.Command) > 0 {
			for m, _ := range v.LivenessProbe.Exec.Command {
				v.LivenessProbe.Exec.Command[m] = strings.Replace(v.LivenessProbe.Exec.Command[m], "MysqlRootPassword", app.Spec.Mysql.MysqlRootPassword, -1)
			}
		}
		if v.ReadinessProbe != nil && v.ReadinessProbe.Exec != nil && len(v.ReadinessProbe.Exec.Command) > 0 {
			for l, _ := range v.ReadinessProbe.Exec.Command {
				v.ReadinessProbe.Exec.Command[l] = strings.Replace(v.ReadinessProbe.Exec.Command[l], "MysqlRootPassword", app.Spec.Mysql.MysqlRootPassword, -1)
			}
		}

	}
	for _, v := range statefulset.Spec.Template.Spec.InitContainers {
		for i, _ := range v.Command {
			v.Command[i] = strings.Replace(v.Command[i], "Mysql-Master-headless", app.Name+"-mysql-0."+app.Name+"-headless", -1)
			v.Command[i] = strings.Replace(v.Command[i], "MysqlRootPassword", app.Spec.Mysql.MysqlRootPassword, -1)
			v.Command[i] = strings.Replace(v.Command[i], "ObjectMeta-Pod-Name", app.Name+"-mysql", -1)
			v.Command[i] = strings.Replace(v.Command[i], "ObjectMeta-Name-headless", app.Name+"-headless", -1)
		}

	}
	if err := controllerutil.SetControllerReference(app, statefulset, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	// 检查是否存在同名的statefulset
	mysqlstate := &appv1.StatefulSet{}

	if err := r.Get(ctx, types.NamespacedName{Name: app.Name + "-mysql", Namespace: app.Namespace}, mysqlstate); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, statefulset); err != nil {
				logger.Error(err, "create mysqlstatefulset failed")
				return ctrl.Result{}, err
			}
		}
	} else {
		if err := r.Update(ctx, statefulset); err != nil {
			return ctrl.Result{}, err
		}
	}
	//4. 处理mycatconfigmap
	mycatconfigmap := utils.NewMycatConfigmap(app)
	if err := controllerutil.SetControllerReference(app, mycatconfigmap, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	//查找同名mycatconfigmap
	mc := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: app.Name + "-mycat-configmap", Namespace: app.Namespace}, mc); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, mycatconfigmap); err != nil {
				logger.Error(err, "create mycatconfigmap failed")
				return ctrl.Result{}, err
			}
		}
	} else {
		if err := r.Update(ctx, mycatconfigmap); err != nil {
			return ctrl.Result{}, err
		}
	}
	//5. 处理mycatdeployment
	mycatdeployment := utils.NewMycatDeploy(app)
	if err := controllerutil.SetControllerReference(app, mycatdeployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	//查找同名mycatdeployment
	dc := &appv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: app.Name + "-mycat", Namespace: app.Namespace}, dc); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, mycatdeployment); err != nil {
				logger.Error(err, "create mycatdeployment failed")
				return ctrl.Result{}, err
			}
		}
	} else {
		if err := r.Update(ctx, mycatdeployment); err != nil {
			return ctrl.Result{}, err
		}
	}
	//5. 处理mycatservice
	mycatservice := utils.NewMycatService(app)

	if err := controllerutil.SetControllerReference(app, mycatservice, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	//查找同名mycatservice
	sc := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: app.Name + "-mycat", Namespace: app.Namespace}, sc); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, mycatservice); err != nil {
				logger.Error(err, "create mycatservice failed")
				return ctrl.Result{}, err
			}
		}
	} else {
		if err := r.Update(ctx, mycatservice); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MysqlrwhaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlv1.Mysqlrwha{}).
		Complete(r)
}
