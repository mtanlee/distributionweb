(function(){
	'use strict';

	angular
		.module('distributionweb.registry')
		.controller('RepositoryController', RepositoryController);

	RepositoryController.$inject = ['resolvedRepository'];
	function RepositoryController(resolvedRepository) {
            var vm = this;
            vm.selectedRepository = resolvedRepository;
	}
})();
