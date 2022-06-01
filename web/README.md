## Environment Prepare

Install `node_modules`:

```bash
npm install
```

or

```bash
yarn
```

## Provided Scripts

Ant Design Pro provides some useful script to help you quick start and build with web project, code style check and test.

Scripts provided in `package.json`. It's safe to modify or add additional script:

### Start project

```bash
npm start
```

### Build project

```bash
npm run build
```

### Check code style

```bash
npm run lint
```

You can also use script to auto fix some lint error:

```bash
npm run lint:fix
```

### Test code

```bash
npm test
```

## 项目部署或更新部署步骤说明

1.执行打包命令，生成一个build文件夹。

2.打开本地电脑终端窗口里，切换到项目文件夹的根目录下，执行如下命令（把`build文件夹` 上传到服务器指定/usr/local/nginx/html目录下），输入服务器密码，部署完成。

```bash
scp -r build root@xxx.xx.xx.xx:/usr/local/nginx/html
```
