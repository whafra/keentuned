import component from './zh-CN/component';
import globalHeader from './zh-CN/globalHeader';
import menu from './zh-CN/menu';
import pwa from './zh-CN/pwa';
import settingDrawer from './zh-CN/settingDrawer';
import settings from './zh-CN/settings';
import pages from './zh-CN/pages';

export default {
  'navBar.lang': '语言',
  'layout.user.link.help': '帮助',
  'layout.user.link.privacy': '隐私',
  'layout.user.link.terms': '条款',
  'app.copyright.produced': '蚂蚁集团体验技术部出品',
  'app.preview.down.block': '下载此页面到本地项目',
  'app.welcome.link.fetch-blocks': '获取全部区块',
  'app.welcome.link.block-list': '基于 block 开发，快速构建标准页面',

  // --------------------
  'operations': '操作',
  'create': '创建',
  'copy': '复制',
  'delete': '删除',
  'edit': '编辑',
  'set': '设置',
  'rollback': '回滚',
  //
  'confirm.title': '删除提示',
  'confirm.content': '确定要删除吗？',
  'btn.yes': '是',
  'btn.no': '否',
  'btn.close': '取消',
  'btn.create': '创建',
  'btn.ok': '确定',
  'btn.confirm': '确认',
  'btn.update': '更新',
  'Input.placeholder': '请输入',
  'Input.rules.message': '仅允许包含字母、数字、下划线、中划线、点，最长32个字符',

  ...pages,
  ...globalHeader,
  ...menu,
  ...settingDrawer,
  ...settings,
  ...pwa,
  ...component,
};
