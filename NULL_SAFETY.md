# 空安全/默认值运算符

本模板引擎现在支持空安全运算符，可以为缺失或空值提供默认值。

## 支持的语法

### 1. 空安全运算符 `??`

```
${variable ?? "默认值"}
```

当 `variable` 为 `nil`、不存在或为空字符串时，返回默认值。

### 2. default 函数

```
${default(variable, "默认值")}
```

功能与 `??` 运算符相同，但使用函数调用语法。

## 使用示例

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${appName ?? "default-app"}
  namespace: ${namespace ?? "default"}
  labels:
    version: ${version ?? "v1.0.0"}
spec:
  replicas: ${replicas ?? 1}
  template:
    spec:
      containers:
        - name: web
          image: ${image ?? "nginx:latest"}
          env:
             - name: LOG_LEVEL
               value: ${logLevel ?? "info"}
```

## 测试结果

运行 `go run main.go` 可以看到以下测试输出：

```
=== 空安全运算符测试 ===

测试结果:
1. 基本 ?? 运算符: 张三
2. 嵌套字段: no-email@example.com
3. default函数: 无描述
4. 数字默认值: 0
5. 空字符串处理: 默认值
```

## 实现细节

- `??` 运算符在预处理阶段被转换为 `default()` 函数调用
- `default()` 函数检查第一个参数是否为 `nil`、不存在或空字符串
- 如果条件满足，返回第二个参数作为默认值
- 支持嵌套字段访问，如 `user.email ?? "default@example.com"`