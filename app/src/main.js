import Vue from 'vue'
import VueRouter from 'vue-router'
import {VueMasonryPlugin} from 'vue-masonry'
import Nav from './Nav.vue'
import Tasks from './Tasks.vue'
import Keyword from './Keyword.vue'
import LaravelVuePagination from 'laravel-vue-pagination';
import Running from './Running.vue'
import Title from './Title.vue'
import Footer from './Footer.vue'

import './style.css'
Vue.use(VueRouter);
Vue.use(VueMasonryPlugin);
Vue.component('pagination', LaravelVuePagination);

const routes = [
    {path: '/', component: Tasks},
    {title: '已启动任务', path: '/tasks', component: Tasks},
    {title: '正在运行', path: '/running-tasks', component: Running},
    {title: '已拒绝任务', path: '/reject-tasks', component: Tasks},
    {title: '关键字', path: '/keywords', component: Keyword},
];

const router = new VueRouter({
    mode: 'history',
    routes
});

new Vue({
    router,
    el: '#app',
    data: {
        routes: routes,
    },
    components: {
        'nav-section': Nav,
        'title-section': Title,
        'footer-section': Footer
    },
});
