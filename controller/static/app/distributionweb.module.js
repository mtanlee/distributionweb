(function(){
    'use strict';

    angular
        .module('distributionweb', [
                'distributionweb.accounts',
                'distributionweb.core',
                'distributionweb.services',
                'distributionweb.layout',
                'distributionweb.help',
                'distributionweb.login',
                'distributionweb.events',
                'distributionweb.registry',
                'distributionweb.filters',
                'angular-jwt',
                'base64',
                'selectize',
                'ui.router'
        ]);

})();
