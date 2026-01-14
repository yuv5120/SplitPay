import { FC } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../components/ui/button';
import { Card, CardContent } from '../components/ui/card';
import { Badge } from '../components/ui/badge';
import { 
  Wallet, 
  Users, 
  TrendingUp, 
  CheckCircle, 
  Zap,
  Shield,
  ArrowRight,
  Receipt
} from 'lucide-react';
import { motion } from 'framer-motion';
import type { LucideIcon } from 'lucide-react';

interface Feature {
  icon: LucideIcon;
  title: string;
  description: string;
}

interface Step {
  number: string;
  title: string;
  description: string;
}

export const LandingPage: FC = () => {
  const navigate = useNavigate();

  const features: Feature[] = [
    {
      icon: Users,
      title: 'Group Management',
      description: 'Create groups for roommates, trips, or any shared expenses'
    },
    {
      icon: Receipt,
      title: 'Track Expenses',
      description: 'Add expenses and split them among participants easily'
    },
    {
      icon: TrendingUp,
      title: 'Smart Calculations',
      description: 'Automatic balance calculations with debt simplification'
    },
    {
      icon: Zap,
      title: 'Lightning Fast',
      description: 'Instant updates and real-time balance tracking'
    },
    {
      icon: Shield,
      title: 'Secure Storage',
      description: 'Your data is safely stored in your browser'
    },
    {
      icon: CheckCircle,
      title: 'Easy Settlements',
      description: 'Minimize transactions with optimal settlement suggestions'
    },
  ];

  const steps: Step[] = [
    {
      number: '01',
      title: 'Create a Group',
      description: 'Start by creating a group and adding members'
    },
    {
      number: '02',
      title: 'Add Expenses',
      description: 'Track who paid for what and who participated'
    },
    {
      number: '03',
      title: 'Settle Up',
      description: 'View balances and optimal settlement suggestions'
    },
  ];

  return (
    <div className="min-h-screen bg-gradient-hero">
      {/* Header */}
      <header className="glass border-b sticky top-0 z-40">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="w-10 h-10 rounded-xl gradient-primary flex items-center justify-center shadow-glow">
                <Wallet className="w-6 h-6 text-primary-foreground" />
              </div>
              <div>
                <h1 className="text-xl font-bold gradient-text">SplitEase</h1>
                <p className="text-xs text-muted-foreground">Smart expense splitting</p>
              </div>
            </div>
            <div className="flex gap-3">
              <Button variant="ghost" onClick={() => navigate('/login')}>
                Sign In
              </Button>
              <Button variant="premium" onClick={() => navigate('/login')}>
                Get Started
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="container mx-auto px-4 py-20 md:py-32">
        <motion.div 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="max-w-4xl mx-auto text-center"
        >
          <Badge variant="glass" className="mb-6 text-sm px-4 py-2">
            <Zap className="w-4 h-4 mr-2" />
            Split expenses easily with friends
          </Badge>
          
          <h1 className="text-5xl md:text-7xl font-bold mb-6 leading-tight">
            Split Expenses
            <br />
            <span className="gradient-text">Effortlessly</span>
          </h1>
          
          <p className="text-xl md:text-2xl text-muted-foreground mb-10 max-w-2xl mx-auto">
            The simplest way to track shared expenses with roommates, friends on trips, or any group activity.
          </p>
          
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button 
              size="lg" 
              variant="premium"
              className="text-lg"
              onClick={() => navigate('/login')}
            >
              Create Your First Group
              <ArrowRight className="w-5 h-5 ml-2" />
            </Button>
            <Button 
              size="lg" 
              variant="outline"
              className="text-lg"
            >
              Learn More
            </Button>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-3 gap-6 mt-16 max-w-2xl mx-auto">
            <div className="text-center">
              <p className="text-4xl font-bold gradient-text">100%</p>
              <p className="text-sm text-muted-foreground mt-1">Free Forever</p>
            </div>
            <div className="text-center">
              <p className="text-4xl font-bold gradient-text">0</p>
              <p className="text-sm text-muted-foreground mt-1">Setup Time</p>
            </div>
            <div className="text-center">
              <p className="text-4xl font-bold gradient-text">∞</p>
              <p className="text-sm text-muted-foreground mt-1">Groups</p>
            </div>
          </div>
        </motion.div>
      </section>

      {/* How It Works */}
      <section className="container mx-auto px-4 py-20 border-t glass">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-4xl md:text-5xl font-bold mb-4">How It Works</h2>
            <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
              Get started in three simple steps. No complicated setup required.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {steps.map((step, idx) => (
              <motion.div
                key={idx}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: idx * 0.1 }}
                viewport={{ once: true }}
              >
                <Card className="glass-card h-full hover:shadow-elegant transition-all group">
                  <CardContent className="p-8">
                    <div className="text-6xl font-bold gradient-text mb-6 opacity-30 group-hover:opacity-100 transition-opacity">
                      {step.number}
                    </div>
                    <h3 className="text-2xl font-bold mb-3 group-hover:gradient-text transition-all">
                      {step.title}
                    </h3>
                    <p className="text-muted-foreground leading-relaxed">
                      {step.description}
                    </p>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="container mx-auto px-4 py-20">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-4xl md:text-5xl font-bold mb-4">Powerful Features</h2>
            <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
              Everything you need to manage shared expenses efficiently
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((feature, idx) => {
              const Icon = feature.icon;
              return (
                <motion.div
                  key={idx}
                  initial={{ opacity: 0, scale: 0.9 }}
                  whileInView={{ opacity: 1, scale: 1 }}
                  transition={{ duration: 0.4, delay: idx * 0.05 }}
                  viewport={{ once: true }}
                >
                  <Card className="glass-card h-full hover:shadow-elegant transition-all group cursor-pointer">
                    <CardContent className="p-6">
                      <div className="w-12 h-12 rounded-lg gradient-primary flex items-center justify-center mb-4 shadow-md group-hover:shadow-glow transition-shadow">
                        <Icon className="w-6 h-6 text-primary-foreground" />
                      </div>
                      <h3 className="text-xl font-semibold mb-2 group-hover:gradient-text transition-all">
                        {feature.title}
                      </h3>
                      <p className="text-muted-foreground leading-relaxed">
                        {feature.description}
                      </p>
                    </CardContent>
                  </Card>
                </motion.div>
              );
            })}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="container mx-auto px-4 py-20">
        <Card className="glass-card max-w-4xl mx-auto">
          <CardContent className="p-12 text-center">
            <h2 className="text-4xl md:text-5xl font-bold mb-6">
              Ready to <span className="gradient-text">Simplify</span> Your Expenses?
            </h2>
            <p className="text-xl text-muted-foreground mb-8 max-w-2xl mx-auto">
              Join thousands of groups managing their shared expenses effortlessly
            </p>
            <Button 
              size="lg" 
              variant="premium"
              className="text-lg"
              onClick={() => navigate('/login')}
            >
              Get Started for Free
              <ArrowRight className="w-5 h-5 ml-2" />
            </Button>
          </CardContent>
        </Card>
      </section>

      {/* Footer */}
      <footer className="border-t glass mt-20">
        <div className="container mx-auto px-4 py-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg gradient-primary flex items-center justify-center">
                <Wallet className="w-5 h-5 text-primary-foreground" />
              </div>
              <span className="font-semibold">SplitEase</span>
            </div>
            <p className="text-sm text-muted-foreground">
              © 2024 SplitEase. Split expenses with friends, simplified.
            </p>
            <div className="flex gap-6 text-sm text-muted-foreground">
              <a href="#" className="hover:text-primary transition-colors">Privacy</a>
              <a href="#" className="hover:text-primary transition-colors">Terms</a>
              <a href="#" className="hover:text-primary transition-colors">Support</a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
};
