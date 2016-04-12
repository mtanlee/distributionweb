(function() {
	'use strict';

	angular
		.module('distributionweb.help')
		.config(getRoutes);

	getRoutes.$inject = ['$stateProvider', '$urlRouterProvider'];

	function getRoutes($stateProvider, $urlRouterProvider) {
		$stateProvider
			.state('dashboard.help', {
				url: '^/help',
				templateUrl: 'app/help/help.html',
                authenticate: true
			});
	}
})();
