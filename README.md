<<<<<<< HEAD
### distributionweb介绍
=======
# distributionweb

>>>>>>> 786c97b0da31692ef410d6e4e30fa65484d100a0
这是个管理docker registry 版本2的web

###项目背景
``这是本人研究和调研docker 容器管理软件过程中接触到shipyard开源项目，并且根据需求对shipyard进行二次开发。shipyard实际上没有实现registry:2的镜像仓库，可以查看registry.go导入的包。本人将开发中的registry：2模块开放出来。这个架构和前端都是根据shipyard和shiyard-ui写的。相关功能在后面会继续完善。```
###开发环境需要的软件

```npm bower docker golang docker-compose```

###部署
```bash
1.下载 git clone https://github.com/mtanlee/distributionweb.git

2.进行项目包的环境配置

3.在distributionweb目录下: docker-compose up 

4.如果没有做dns，添加hostname，需要更新config/domain.crt，使用你的registry的domain.crt
  docker exec -it distribution_controller_1 bash echo "ip_address hostname" >> /etc/hosts
```
