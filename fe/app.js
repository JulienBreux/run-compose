var app = angular.module('mealApp', []);

app.controller('MealController', function($scope, $http) {
    $scope.meals = [];
    $scope.newMeal = {};

    $scope.loadMeals = function() {
        $http.get('/api/meals').then(function(response) {
            $scope.meals = response.data || [];
        }, function(error) {
            console.error('Error loading meals:', error);
        });
    };

    $scope.addMeal = function() {
        // Convert calories to int
        $scope.newMeal.calories = parseInt($scope.newMeal.calories, 10);

        $http.post('/api/meals', $scope.newMeal).then(function(response) {
            $scope.newMeal = {}; // Reset form
            $scope.loadMeals(); // Reload list
        }, function(error) {
            console.error('Error adding meal:', error);
        });
    };

    // Initial load
    $scope.loadMeals();
});
