import menu from './en-US/menu';
import pages from './en-US/pages';
import pwa from './en-US/pwa';
import settingDrawer from './en-US/settingDrawer';

export default {
  'navBar.lang': 'Languages',
  'layout.user.link.help': 'Help',
  'layout.user.link.privacy': 'Privacy',
  'layout.user.link.terms': 'Terms',

  // --------------------
  'operations': 'Operations',
  'create': 'Create',
  'copy': 'Copy',
  'delete': 'Delete',
  'edit': 'Edit',
  'set': 'Set',
  'rollback': 'Rollback',
  'rerun': 'Rerun',
  //
  'confirm.title': 'Delete Tips',
  'confirm.content': 'Are you sure you want to delete?',
  'btn.yes': 'Yes',
  'btn.no': 'No',
  'btn.close': 'Cancel',
  'btn.create': 'Create',
  'btn.ok': 'Ok',
  'btn.confirm': 'Confirm',
  'btn.update': 'Update',

  'Input.placeholder': 'Please input',
  'Input.rules.message': 'Only letters, numbers, underscores, middle dashes and dots are allowed. The maximum length is 32 characters',
  // 分页
  'total': ' ',
  'records': 'records page',
  'page': ' ',

    ...menu,
    ...pages,
    ...settingDrawer,
    ...pwa,
};
