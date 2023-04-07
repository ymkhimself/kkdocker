# kkdocker

一个简易的容器引擎,Go开发，基于Namespace,cgroup,aufs。使用bridge
支持命令:
- [run](run.go):启动容器，可后台运行，可挂载volume，可限制容器资源
- [stop](stop.go):停止容器
- [rm](remove.go):删除容器
- [ps](list.go):列出容器
- [logs](log.go):显示容器日志
- [commit](commit.go):提交容器，打成镜像
- [exec](exec.go):进入容器执行命令