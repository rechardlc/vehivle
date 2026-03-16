# internal/service/param_template

## 作用
参数模板与参数项维护

## 业务范围
模板 CRUD、参数项排序、必填与展示位控制。

## 关键规则
模板绑定一级分类、空值展示规则由模板驱动。

## 学习建议
先把该域的领域规则写清楚，再实现 repository 接口，最后接 handler。
