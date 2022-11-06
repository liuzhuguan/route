package service

import (
	"context"
	"git.imooc.com/coding-535/common"
	"github.com/liuzhuguan/route/domain/model"
	"github.com/liuzhuguan/route/domain/repository"
	"github.com/liuzhuguan/route/proto/route"
	"k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"

	v12 "k8s.io/api/networking/v1"
	v14 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//这里是接口类型
type IRouteDataService interface {
	AddRoute(*model.Route) (int64, error)
	DeleteRoute(int64) error
	UpdateRoute(*model.Route) error
	FindRouteByID(int64) (*model.Route, error)
	FindAllRoute() ([]model.Route, error)

	CreateRouteToK8s(*route.RouteInfo) error
	DeleteRouteFromK8s(*model.Route) error
	UpdateRouteToK8s(*route.RouteInfo) error
}

//创建
//注意：返回值 IRouteDataService 接口类型
func NewRouteDataService(routeRepository repository.IRouteRepository, clientSet *kubernetes.Clientset) IRouteDataService {
	return &RouteDataService{RouteRepository: routeRepository, K8sClientSet: clientSet, deployment: &v1.Deployment{}}
}

type RouteDataService struct {
	//注意：这里是 IRouteRepository 类型
	RouteRepository repository.IRouteRepository
	K8sClientSet    *kubernetes.Clientset
	deployment      *v1.Deployment
}

//插入
func (u *RouteDataService) AddRoute(route *model.Route) (int64, error) {
	return u.RouteRepository.CreateRoute(route)
}

//删除
func (u *RouteDataService) DeleteRoute(routeID int64) error {
	return u.RouteRepository.DeleteRouteByID(routeID)
}

//更新
func (u *RouteDataService) UpdateRoute(route *model.Route) error {
	return u.RouteRepository.UpdateRoute(route)
}

//查找
func (u *RouteDataService) FindRouteByID(routeID int64) (*model.Route, error) {
	return u.RouteRepository.FindRouteByID(routeID)
}

//查找
func (u *RouteDataService) FindAllRoute() ([]model.Route, error) {
	return u.RouteRepository.FindAll()
}

func (u *RouteDataService) CreateRouteToK8s(routeInfo *route.RouteInfo) error {
	ingress := u.setIngress(routeInfo)
	// 查找是否存在，不存在创建
	if _, err := u.K8sClientSet.NetworkingV1().Ingresses(routeInfo.RouteNamespace).Get(context.TODO(), routeInfo.RouteName, v14.GetOptions{}); err != nil {
		if _, err := u.K8sClientSet.NetworkingV1().Ingresses(routeInfo.RouteNamespace).Create(context.TODO(), ingress, v14.CreateOptions{}); err != nil {
			common.Error("CreateRouteToK8s create err: ", err.Error())
			return err
		} else {
			return nil
		}
	} else {
		common.Error(routeInfo.RouteNamespace + "-" + routeInfo.RouteName + "exist")
		return nil
	}
}

func (u *RouteDataService) DeleteRouteFromK8s(r *model.Route) error {
	// 查找是否存在，不存在创建
	if _, err := u.K8sClientSet.NetworkingV1().Ingresses(r.RouteNamespace).Get(context.TODO(), r.RouteName, v14.GetOptions{}); err == nil {
		if err := u.K8sClientSet.NetworkingV1().Ingresses(r.RouteNamespace).Delete(context.TODO(), r.RouteName, v14.DeleteOptions{}); err != nil {
			common.Error("CreateRouteToK8s create err: ", err.Error())
			return err
		} else {
			return nil
		}
	} else {
		common.Error(r.RouteNamespace + "-" + r.RouteName + " not exist")
		return err
	}
}

func (u *RouteDataService) UpdateRouteToK8s(routeInfo *route.RouteInfo) error {
	ingress := u.setIngress(routeInfo)

	if _, err := u.K8sClientSet.NetworkingV1().Ingresses(routeInfo.RouteNamespace).Get(context.TODO(), routeInfo.RouteName, v14.GetOptions{}); err == nil {
		if _, err := u.K8sClientSet.NetworkingV1().Ingresses(routeInfo.RouteNamespace).Update(context.TODO(), ingress, v14.UpdateOptions{}); err != nil {
			common.Error("CreateRouteToK8s create err: ", err.Error())
			return err
		} else {
			return nil
		}
	} else {
		common.Error(routeInfo.RouteNamespace + "-" + routeInfo.RouteName + " not exist")
		return err
	}
}

func (u *RouteDataService) setIngress(info *route.RouteInfo) *v12.Ingress {
	r := &v12.Ingress{}
	// set route
	r.TypeMeta = v14.TypeMeta{
		Kind:       "Ingress",
		APIVersion: "1.0",
	}
	// set route info
	r.ObjectMeta = v14.ObjectMeta{
		Name:      info.RouteName,
		Namespace: info.RouteNamespace,
		Labels: map[string]string{
			"app-name": info.RouteName,
			"author":   "lzg",
		},
		Annotations: map[string]string{
			"k8s/generated-by-lzg": "由lzg老师代码创建",
		},
	}
	//使用 ingress-nginx
	className := "nginx"
	//设置路由 spec 信息
	r.Spec = v12.IngressSpec{
		IngressClassName: &className,
		//默认访问服务
		DefaultBackend: nil,
		//如果开启https这里要设置
		TLS:   nil,
		Rules: u.getIngressPath(info),
	}
	return r
}

func (u *RouteDataService) getIngressPath(info *route.RouteInfo) (path []v12.IngressRule) {
	//1.设置host
	pathRule := v12.IngressRule{Host: info.RouteHost}
	//2.设置Path
	ingressPath := []v12.HTTPIngressPath{}
	for _, v := range info.RoutePath {
		pathType := v12.PathTypePrefix
		ingressPath = append(ingressPath, v12.HTTPIngressPath{
			Path:     v.RoutePathName,
			PathType: &pathType,
			Backend: v12.IngressBackend{
				Service: &v12.IngressServiceBackend{
					Name: v.RouteBackendService,
					Port: v12.ServiceBackendPort{
						Number: v.RouteBackendServicePort,
					},
				},
			},
		})
	}

	//3.赋值 Path
	pathRule.IngressRuleValue = v12.IngressRuleValue{HTTP: &v12.HTTPIngressRuleValue{Paths: ingressPath}}
	path = append(path, pathRule)
	return
}
