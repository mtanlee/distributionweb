(function(){
	'use strict';

	angular
		.module('distributionweb.events')
        .factory('EventsService', EventsService);

	EventsService.$inject = ['$resource'];
	function EventsService($resource) {
            return $resource('/api/events');
	}
})();
