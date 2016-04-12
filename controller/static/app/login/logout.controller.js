(function(){
	'use strict';

	angular
		.module('distributionweb.login')
		.controller('LogoutController', LogoutController);

    LogoutController.$inject = ['AuthService', '$state'];
	function LogoutController(AuthService, $state) {
            var vm = this;
            AuthService.logout();
            $state.transitionTo('login');
        }
})();

