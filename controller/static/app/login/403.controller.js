(function(){
	'use strict';

	angular
	    .module('distributionweb.login')
	    .controller('AccessDeniedController', AccessDeniedController);

	AccessDeniedController.$inject = ['$stateParams'];
	function AccessDeniedController($stateParams) {
            var vm = this;
	}
})();
