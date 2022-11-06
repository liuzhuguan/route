package handler

import (
	"context"
	"git.imooc.com/coding-535/common"
	log "github.com/asim/go-micro/v3/logger"
	"github.com/liuzhuguan/route/domain/model"
	"github.com/liuzhuguan/route/domain/service"
	"github.com/liuzhuguan/route/proto/route"
	"github.com/pkg/errors"
	"strconv"
)

type RouteHandler struct {
	//注意这里的类型是 IRouteDataService 接口类型
	RouteDataService service.IRouteDataService
}

// Call is a single request handler called via client.Call or the generated client code
func (e *RouteHandler) AddRoute(ctx context.Context, info *route.RouteInfo, rsp *route.Response) error {
	log.Info("Received *route.AddRoute request: ", info)
	r := &model.Route{}
	if err := common.SwapTo(info, r); err != nil {
		common.Error(err)
		rsp.Msg = err.Error()
		return err
	}
	//创建route到k8s
	if err := e.RouteDataService.CreateRouteToK8s(info); err != nil {
		common.Error(err)
		rsp.Msg = err.Error()
		return err
	} else {
		//写入数据库
		routeID, err := e.RouteDataService.AddRoute(r)
		if err != nil {
			common.Error(err)
			rsp.Msg = err.Error()
			return err
		}
		common.Info("Route 添加成功 ID 号为：" + strconv.FormatInt(routeID, 10))
		rsp.Msg = "Route 添加成功 ID 号为：" + strconv.FormatInt(routeID, 10)
	}
	return nil
}

func (e *RouteHandler) DeleteRoute(ctx context.Context, req *route.RouteId, rsp *route.Response) error {
	log.Info("Received *route.DeleteRoute request: ", req)

	r, err := e.RouteDataService.FindRouteByID(req.Id)
	if err != nil {
		common.Error(err)
		return err
	}

	if err = e.RouteDataService.DeleteRouteFromK8s(r); err != nil {
		common.Error(err)
		rsp.Msg = err.Error()
		return err
	} else {
		//写入数据库
		err = e.RouteDataService.DeleteRoute(req.Id)
		if err != nil {
			common.Error(err)
			rsp.Msg = err.Error()
			return err
		}
	}
	return nil
}

func (e *RouteHandler) UpdateRoute(ctx context.Context, req *route.RouteInfo, rsp *route.Response) error {
	log.Info("Received *route.UpdateRoute request: ", req)

	r, err := e.RouteDataService.FindRouteByID(req.Id)
	if err != nil || r == nil {
		common.Error("UpdateRoute err, route info is nil or err:", err)
		return errors.New("UpdateRoute err, route info is nil or err:" + err.Error())
	}

	newRoute := &model.Route{}
	if err := common.SwapTo(req, newRoute); err != nil {
		common.Error(err)
		rsp.Msg = err.Error()
		return err
	}

	if err = e.RouteDataService.UpdateRouteToK8s(req); err != nil {
		common.Error(err)
		rsp.Msg = err.Error()
		return err
	} else {
		//写入数据库
		err = e.RouteDataService.UpdateRoute(newRoute)
		if err != nil {
			common.Error(err)
			rsp.Msg = err.Error()
			return err
		}
	}
	return nil
}

func (e *RouteHandler) FindRouteByID(ctx context.Context, req *route.RouteId, rsp *route.RouteInfo) error {
	log.Info("Received *route.FindRouteByID request: ", req)

	r, err := e.RouteDataService.FindRouteByID(req.Id)
	if err != nil || r == nil {
		common.Error("UpdateRoute err, route info is nil or err:", err)
		return errors.New("UpdateRoute err, route info is nil or err:" + err.Error())
	}

	//数据转化
	if err = common.SwapTo(r, rsp); err != nil {
		common.Error(err)
		return err
	}
	return nil
}

func (e *RouteHandler) FindAllRoute(ctx context.Context, req *route.FindAll, rsp *route.AllRoute) error {
	log.Info("Received *route.FindAllRoute request")
	allRoute, err := e.RouteDataService.FindAllRoute()
	if err != nil {
		common.Error(err)
		return err
	}
	//整理下格式
	for _, v := range allRoute {
		//创建实例
		routeInfo := &route.RouteInfo{}
		//把查询出来的数据进行转化
		if err := common.SwapTo(v, routeInfo); err != nil {
			common.Error(err)
			return err
		}
		//数据合并
		rsp.RouteInfo = append(rsp.RouteInfo, routeInfo)
	}
	return nil
}
