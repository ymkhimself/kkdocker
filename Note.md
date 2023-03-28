# 如何实现容器的后台运行
早期的Docker，所有容器的init进程都是从docker daemon这个进程fork出来的 ，这也就会导致，如果docker daemon挂到，所有容器都G了。后来docker使用了
containerd，low-level的实现是runC，可以实现daemon挂掉，容器仍然在。

父进程退出子进程不挂的原理：命令mydocker是主进程，创建出的子进程是我们的容器，如果父进程挂了，子进程就成了孤儿进程。为了避免无法释放他占用的资源，进程号
为1的进程init就会接受这些孤儿进程。
> 疑惑，父进程退出，子进程到底是个啥状态，前后说的好像不一样呀

# 如何实现docker ps
非常简单，Docker启动时，我们要生成docker信息，放在/var/run/mydocker 下，一个容器一个文件夹，容器要有名字

文件夹下的config.json文件储存容器的信息，包括id,pid,status,command,created等

docker ps的时候，去遍历/var/run/mydocker目录就行了

# 如何实现 docker logs
将docker容器内的标准输出重定向到一个文件，就ok了


# 如何实现 docker exec
容器创建之后，我们就无法再进入容器了，我们要怎么再次进入容器呢，这时候需要再次进入容器的namespace

## setns
setns 是一个系统调用，可以根据提供的pid再次进入指定的namespace。
但是对于go来说很麻烦，对于Mount Namespace来说，一个具有多线程的进程无法使用setns调用进入到对应的命名空间。但是Go每启动一个程序都会进入多线程
状态，因此无法简单地通过go调用这个系统调用，需要用C

## Cgo
Cgo允许Go程序去调用C的函数和标准库。你只需要以一种特殊的方式在Go的源码里面写出需要调用的c代码，Cgo就能把你的C代码和Go文件整合成一个包。

# 如何停止和删除容器
## 停止
1. 根据容器名获取pid
2. kill pid
3. 修改容器信息
4. 将容器信息写回去
## 删除
1. 根据容器名获取容器信息
2. 判断容器状态
3. 删除容器信息文件

# 通过容器制作镜像
