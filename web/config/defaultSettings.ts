import { Settings as LayoutSettings } from '@ant-design/pro-layout';

const Settings: LayoutSettings & {
  pwa?: boolean;
  logo?: string;
} = {
  navTheme: 'light',
  primaryColor: '#1890ff', // 拂晓蓝
  layout: 'top',//'mix',
  contentWidth: 'Fluid',
  fixedHeader: false,
  fixSiderbar: true,
  colorWeak: false,
  title: 'KEENTUNE',
  pwa: false,
  logo: '',
  iconfontUrl: '',
};

export default Settings;
