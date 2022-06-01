export default [
  {
    path: '/home',
    name: 'home',
    layout: false,
    hideInMenu: true,
    component: './Home',
  },
  {
    path: '/list',
    name: 'list',
    hideInMenu: true,
    component: './List',
      routes: [
        {
          path: '/list/static-page',
          name: 'static-page',
          icon: 'smile',
          component: './List/modules/ProfileList',
        },
        // {
        //   path: '/list/sensitive-parameter',
        //   name: 'sensitive-parameter',
        //   icon: 'smile',
        //   component: './List/modules/SensitivityTaskList',
        // },
        // {
        //   path: '/list/sensitive-parameter/details',
        //   name: 'sensitive-parameter-details',
        //   icon: 'smile',
        //   component: './List/modules/SensitivityTaskList/Details',
        // },
        // {
        //   path: '/list/tuning-task',
        //   name: 'tuning-task',
        //   component: './List/modules/TuningTaskList',
        // },
        // {
        //   path: '/list/tuning-task/details',
        //   name: 'tuning-task-details',
        //   component: './List/modules/TuningTaskList/Details',
        // },

        // --------------------------------------------------
        // {
        //   path: '/list/sensitivity/compare',
        //   name: 'sensitivity-compare',
        //   icon: 'smile',
        //   component: './List/modules/SensitivityTaskList/Compare',
        // },
        // {
        //   path: '/list/tuning-task/compare',
        //   name: 'tuning-task-compare',
        //   icon: 'smile',
        //   component: './List/modules/TuningTaskList/Compare',
        // },
        {
          component: './404',
        },
      ]
  },

  {
    path: '/',
    redirect: '/home',
  },
  {
    component: './404',
  },
];
