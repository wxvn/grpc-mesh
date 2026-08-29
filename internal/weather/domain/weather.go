package domain

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

type WeatherCondition string

const (
	ConditionClear            WeatherCondition = "clear"
	ConditionMostlyClear      WeatherCondition = "mostly_clear"
	ConditionPartlyCloudy     WeatherCondition = "partly_cloudy"
	ConditionOvercast         WeatherCondition = "overcast"
	ConditionFog              WeatherCondition = "fog"
	ConditionDrizzle          WeatherCondition = "drizzle"
	ConditionFreezingDrizzle  WeatherCondition = "freezing_drizzle"
	ConditionRain             WeatherCondition = "rain"
	ConditionFreezingRain     WeatherCondition = "freezing_rain"
	ConditionSnow             WeatherCondition = "snow"
	ConditionSnowGrains       WeatherCondition = "snow_grains"
	ConditionRainShowers      WeatherCondition = "rain_showers"
	ConditionSnowShowers      WeatherCondition = "snow_showers"
	ConditionThunderstorm     WeatherCondition = "thunderstorm"
	ConditionThunderstormHail WeatherCondition = "thunderstorm_hail"
	ConditionUnknown          WeatherCondition = "unknown"
)

type CurrentWeather struct {
	City        string
	Coordinates Coordinates
	Elevation   float64

	Temperature float64
	FeelsLike   float64
	Humidity    int32
	WindSpeed   float64

	Condition WeatherCondition
	Time      string
	Timezone  string
}

type ForecastDay struct {
	Date                     string
	TempMin                  float64
	TempMax                  float64
	PrecipitationProbability int32
	Condition                WeatherCondition
}

type Forecast struct {
	City        string
	Coordinates Coordinates
	Elevation   float64
	Days        []ForecastDay
}
