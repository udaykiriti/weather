package weather

import (
	"math/rand"
	"strings"
)

// Quote returns a funny weather quote matching the WMO code.
func Quote(wmoCode int) string {
	var pool []string
	switch {
	case wmoCode == 0:
		pool = []string{
			"Sun's out, bad decisions are out too.",
			"Perfect weather to pretend you're a lizard on a rock.",
			"The sky is blue. Your excuses are not.",
			"Vitamin D loading... please wait.",
			"It's so sunny even your shadow needs sunglasses.",
		}
	case wmoCode == 1:
		pool = []string{
			"Mostly clear — like your schedule should be.",
			"A few clouds, just enough to keep the sky humble.",
			"The sun is trying its best. Same energy.",
		}
	case wmoCode == 2:
		pool = []string{
			"Partly cloudy, fully indecisive.",
			"The weather can't make up its mind. Neither can you. Perfect match.",
			"Clouds auditioning for a role in your afternoon plans.",
		}
	case wmoCode == 3:
		pool = []string{
			"Overcast. Great day to feel dramatically misunderstood.",
			"Zero sun, maximum brooding potential.",
			"The sky is wearing a grey blanket. Take notes.",
			"Overcast skies: nature's way of saying 'meh'.",
		}
	case wmoCode == 45 || wmoCode == 48:
		pool = []string{
			"Fog warning: if you can't see your problems, do they even exist?",
			"It's foggy. Perfect alibi weather.",
			"Visibility low. Mystery high.",
			"Great day to dramatically disappear into the mist.",
		}
	case wmoCode >= 51 && wmoCode <= 55:
		pool = []string{
			"Drizzle. Nature's way of passive-aggressively watering your plans.",
			"It's not raining, it's misting. Like a fancy spa you didn't ask for.",
			"Light drizzle: too wet to ignore, too weak to respect.",
		}
	case wmoCode == 61:
		pool = []string{
			"Slight rain. A solid excuse not to go jogging.",
			"Rain check? The sky literally issued one.",
			"Nature is crying. Relatable.",
		}
	case wmoCode == 63:
		pool = []string{
			"Moderate rain. Your hair has accepted its fate.",
			"It's raining. Cancel everything and make soup.",
			"Rain: nature's way of doing your car wash for free.",
		}
	case wmoCode == 65:
		pool = []string{
			"Heavy rain. You ARE the soup now.",
			"It's pouring. Even the ducks are impressed.",
			"Biblical rain detected. Start building something.",
			"Congratulations, you're basically underwater.",
		}
	case wmoCode >= 71 && wmoCode <= 75:
		pool = []string{
			"Snow! Nature said 'let me delete everything and start fresh'.",
			"It's snowing. Time to question every life choice that led you here.",
			"Snow: beautiful from inside. Terrible from outside.",
			"White stuff everywhere. And it's not sugar.",
		}
	case wmoCode == 77:
		pool = []string{
			"Snow grains. Tiny frozen disappointments falling from the sky.",
			"Snow grains: the economy-sized version of hail.",
		}
	case wmoCode >= 80 && wmoCode <= 82:
		pool = []string{
			"Rain showers incoming. The sky has commitment issues.",
			"On-and-off rain. Like a bad situationship.",
			"Showers: enough rain to ruin your day, not enough to cancel plans.",
		}
	case wmoCode >= 85 && wmoCode <= 86:
		pool = []string{
			"Snow showers. Nature's confetti, but colder.",
			"It's snowing intermittently, like inspiration.",
		}
	case wmoCode == 95:
		pool = []string{
			"Thunderstorm. Nature is having a moment.",
			"Lightning detected. Unplug your WiFi router and panic.",
			"Thor is upset about something. As usual.",
			"Great day to feel small and insignificant. Nature's doing the work.",
		}
	case wmoCode == 96 || wmoCode == 99:
		pool = []string{
			"Thunderstorm with hail. Nature said 'not today'.",
			"Hail + lightning. Your car's worst nightmare.",
			"The sky is literally throwing rocks at you. Take the hint and stay inside.",
		}
	default:
		pool = []string{
			"Weather: it exists. Outside: also exists. You: reading this.",
			"Conditions unknown. Like your weekend plans.",
		}
	}
	return pool[rand.Intn(len(pool))]
}

// Advice returns a feels-like temperature advice string.
func Advice(feelsLike float64, unit string) string {
	// Normalise to Celsius for comparison
	temp := feelsLike
	if unit == "°F" {
		temp = (feelsLike - 32) * 5 / 9
	}

	switch {
	case temp <= -20:
		return "It feels arctic out there. Wrap up like a burrito."
	case temp <= -10:
		return "Dangerously cold. Only go outside if your name is a penguin."
	case temp <= 0:
		return "Below freezing. Every exposed inch of skin will regret this."
	case temp <= 5:
		return "Heavy coat mandatory. Your nose will run regardless."
	case temp <= 10:
		return "Jacket weather. The kind that makes you question the seasons."
	case temp <= 15:
		return "A light jacket will do. Maybe two. Bring both."
	case temp <= 20:
		return "Comfortable. Wear what you want, nobody's judging."
	case temp <= 25:
		return "T-shirt weather. Go enjoy it — you earned this."
	case temp <= 30:
		return "Warm. Stay hydrated and pretend you love summer."
	case temp <= 35:
		return "Hot. Ice cream is not optional at this point."
	case temp <= 40:
		return "Dangerously hot. You are now a human crouton."
	default:
		return "It's basically an oven outside. Stay in. Order food."
	}
}

// QuoteFromIcon picks a funny quote by mapping the weather-icons CSS class back
// to a condition bucket, so the web server doesn't need to pass raw WMO codes.
func QuoteFromIcon(iconClass string) string {
	switch {
	case strings.Contains(iconClass, "sunny") && !strings.Contains(iconClass, "overcast"):
		return Quote(0)
	case strings.Contains(iconClass, "sunny-overcast"):
		return Quote(1)
	case strings.Contains(iconClass, "cloudy") && strings.Contains(iconClass, "day"):
		return Quote(2)
	case strings.Contains(iconClass, "cloudy"):
		return Quote(3)
	case strings.Contains(iconClass, "fog"):
		return Quote(45)
	case strings.Contains(iconClass, "sprinkle") || strings.Contains(iconClass, "rain-mix"):
		return Quote(51)
	case strings.Contains(iconClass, "rain-wind"):
		return Quote(65)
	case strings.Contains(iconClass, "rain"):
		return Quote(63)
	case strings.Contains(iconClass, "snow-wind"):
		return Quote(75)
	case strings.Contains(iconClass, "snowflake"):
		return Quote(77)
	case strings.Contains(iconClass, "snow"):
		return Quote(71)
	case strings.Contains(iconClass, "showers"):
		return Quote(80)
	case strings.Contains(iconClass, "storm-showers"):
		return Quote(99)
	case strings.Contains(iconClass, "thunderstorm"):
		return Quote(95)
	default:
		return Quote(-1)
	}
}
