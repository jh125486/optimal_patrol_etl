library(ggplot2)
library(rgdal)
library(ggmap)
library(scales)

crimes <- read.csv("./results/crimes.csv")
# remove crimes with weight 0 and are outside normal area
crimes = crimes[which(crimes$Weight > 0 & crimes$X >= 6270000 & crimes$X <= 6680000),]
# replicate all crimes by weight
crimes <- crimes[rep(1:nrow(crimes), crimes[,3]), 5:6]
crimes$Weight <- factor(crimes$Weight)
crimes$Gang <- NULL   
scale = 2
width = 6.5
heatcols <- heat.colors(5)
heatcols = rev(heatcols)
lamap <- get_map(location = 'LA County', source = 'google', maptype = 'roadmap', color= "bw")
weekdays = c("Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday")


# plot weights
for (i in 1:max(crimes$Weight)) {
    subcrimes = subset(crimes, Weight == i)
    clustered = kmeans(subcrimes[5:6], 50) 
    subcrimes$cluster = factor(clustered$cluster)
    centers=as.data.frame(clustered$centers)
    
    filename = paste("cluster_weights_", i, ".png", sep = "")
    ggplot(data=subcrimes, aes(X, Y, colour=cluster)) + 
      geom_point(alpha = 1/10) + 
      geom_point(data=centers, aes(X, Y, colour='Center'), color="black") + 
      labs(title = sprintf("Crimes: weight %d @ 50 clusters", i)) + 
      coord_fixed()
    ggsave(file=filename, scale=scale, width=width) 
} 
# plot hours
for (i in min(crimes$Hour):max(crimes$Hour)) {
    subcrimes = subset(crimes, Hour == i)
    clustered = kmeans(subcrimes[5:6], 50) 
    subcrimes$cluster = factor(clustered$cluster)
    centers=as.data.frame(clustered$centers)
    
    filename = paste("cluster_hour_", i, ".png", sep = "")
    ggplot(data=subcrimes, aes(X, Y, colour=cluster)) + 
      geom_point(alpha = 1/10) + 
      geom_point(data=centers, aes(X, Y, colour='Center'), color="black") + 
      labs(title = sprintf("Crimes: hour %d @ 50 clusters", i)) + 
      coord_fixed()
    ggsave(file=filename, scale=scale, width=width)
}

# plot Days of week
for (i in min(crimes$DoW):max(crimes$DoW)) {
    subcrimes = subset(crimes, Hour == i)
    clustered = kmeans(subcrimes[5:6], 50) plot
    subcrimes$cluster = factor(clustered$cluster)
    centers=as.data.frame(clustered$centers)
    
    weekday = weekdays[i+1][1]
    filename = paste("cluster_", weekday, ".png", sep = "") 
    ggplot(data=subcrimes, aes(X, Y, colour=cluster)) + 
      geom_point(alpha = 1/10) + 
      geom_point(data=centers, aes(X, Y, colour='Center'), color="black") + 
      labs(title = sprintf("Crimes: %s @ 50 clusters", weekday)) + 
      coord_fixed()
    ggsave(file=filename, scale=scale, width=width)
}

# do overall cluster
clustered = kmeans(subcrimes[5:6], 50, nstart=10) 
subcrimes$cluster = factor(clustered$cluster)
centers=as.data.frame(clustered$centers)
coordinates(centers) <- ~ X + Y
proj4string(centers) <- CRS("+init=epsg:2229")
spTransform(centers, CRS("+init=epsg:4326"))

# plot overall clusters
filename = paste("clusters_50.png", sep = "")
ggplot(data=subcrimes, aes(X, Y, colour=cluster)) + 
    geom_point(alpha = 1/10) + 
    geom_point(data=centers, aes(X,Y, colour='Center'), color="black") +
    labs(title = "All crimes @ 50 clusters") + 
    coord_fixed()
ggsave(file=filename, scale=scale, width=width) 

coordinates(centers) <- ~ X + Y
proj4string(centers) <- CRS("+init=epsg:2229")

centroids = spTransform(centers, CRS("+init=epsg:4326"))
centroids=as.data.frame(centroids)
qmap('Los Angeles County', zoom=9, maptype='roadmap', color='bw') + 
    geom_point(data=centroids, aes(x = X, y = Y), color="red", size=2) + 
    geom_label(data=centroids, aes(x=X, y=Y, label=sprintf("%f, %f", Y, X)), size = 3, nudge_y = -0.01, nudge_x = 0.08)
    


qmap('Los Angeles County', zoom=10, maptype='roadmap', color='bw') + 
	geom_point(data=nodes, aes(y = Latitude, x = Longitude, size= Weight), color="red")
    

################################################################################
# set font size
theme_set(theme_gray(base_size = 18))

# histogram of crimes
ggplot(data=by_day, aes(x=Day, y=Crimes, fill=-Crimes)) + 
	geom_bar(stat="identity", color="black") + 
	scale_y_log10(expand=c(0,0), breaks=trans_breaks("log10", function(x) 10^x), labels=trans_format("log10", math_format(10^.x))) + 
	scale_x_discrete(name="Weekday", limits=weekdays) + 
	annotation_logticks(sides='l') +
	guides(fill=FALSE) 
	
ggplot(data=by_weight, aes(x=Weight, y=Crimes, fill=factor(Weight))) + 
	geom_bar(stat="identity", color="black") + 
	scale_y_log10(expand=c(0,0), breaks=trans_breaks("log10", function(x) 10^x), labels=trans_format("log10", math_format(10^.x))) + 
	annotation_logticks(sides='l') + 
	scale_fill_manual(values=heatcols) + 
	guides(fill=FALSE)
	
by_hour$Hour = factor(by_hour$Hour)
ggplot(data=by_hour, aes(x=Hour, y= Crimes, fill=-Crimes)) + 
	geom_bar(stat="identity", color="black") + 
	scale_y_log10(expand=c(0,0), breaks=trans_breaks("log10", function(x) 10^x), labels=trans_format("log10", math_format(10^.x))) +  
	scale_x_discrete("Hour", breaks=0:23,labels=c("00","01","02", "03","04","05","06","07","08","09","10","11","12","13","14","15","16","17","18","19","20","21","22","23")) + 
	annotation_logticks(sides='l') + 
	guides(fill=FALSE)

# plot overall
ggmap(lamap) + 
    geom_point(data=crimes, aes(X, Y, color=Weight), alpha = 1/10) +
    scale_color_manual(values=heatcols) + 
    guides(color = guide_legend(override.aes = list(alpha = 1))) + 
    ggtitle(sprintf("All crimes", i))
ggsave(file=sprintf("all.png", i), scale=scale, width=width) 
ggmap(lamap) + 
    stat_density2d(aes(X, Y, fill=Weight, alpha=..level..), size=1, bins=128, data=crimes, geom="polygon") + 
    scale_fill_manual(values=heatcols) +
    ggtitle(sprintf("All crime density", i))
ggsave(file=sprintf("all_density.png", i), scale=scale, width=width)

# plot hours
for (i in min(crimes$Hour):max(crimes$Hour)) {
    subcrimes = subset(crimes, Hour == i)
    print(nrow(subcrimes))
    ggmap(lamap) + 
        geom_point(data=subcrimes, aes(X, Y, color=Weight), alpha = 1/10) +
        scale_color_manual(values=heatcols) + 
        guides(color = guide_legend(override.aes = list(alpha = 1))) + 
        ggtitle(sprintf("Hour %02d00 crimes", i))
    ggsave(file=sprintf("%02d00.png", i), scale=scale, width=width) 
    
    ggmap(lamap) + 
        stat_density2d(aes(X, Y, fill=Weight, alpha=..level..), size=1, bins=128, data=subcrimes, geom="polygon") + 
        scale_fill_manual(values=heatcols) +
        ggtitle(sprintf("Hour %02d00 crime density", i))
    ggsave(file=sprintf("%02d00_density.png", i), scale=scale, width=width)
}

# plot days
for (i in min(crimes$DoW):max(crimes$DoW)) {
    weekday = weekdays[i+1][1]
    subcrimes = subset(crimes, DoW == i)
    print(nrow(subcrimes))
    ggmap(lamap) + 
        geom_point(data=subcrimes, aes(X, Y, color=Weight), alpha = 1/10) +
        scale_color_manual(values=heatcols) + 
        guides(color = guide_legend(override.aes = list(alpha = 1))) + 
        ggtitle(sprintf("%s crimes", weekday))
    ggsave(file=sprintf("%s.png", weekday), scale=scale, width=width) 
    
    ggmap(lamap) + 
        stat_density2d(aes(X, Y, fill=Weight, alpha=..level..), size=1, bins=128, data=subcrimes, geom="polygon") +
        ggtitle(sprintf("%s crime density", i)) +
        scale_fill_manual(values=heatcols)
    ggsave(file=sprintf("%s_density.png", weekday), scale=scale, width=width)
}

# plot weights
for (i in "1":"5") {
    subcrimes = subset(crimes, Weight == i)
    print(nrow(subcrimes))
    ggmap(lamap) + 
        geom_point(data=subcrimes, aes(X, Y, color=heatcols(Weight)), alpha = 1/10) +
        guides(color=FALSE) +
        ggtitle(sprintf("Crimes by weight %d", i))
    ggsave(file=sprintf("weight_%s.png", i), scale=scale, width=width) 
    
    ggmap(lamap) + 
        stat_density2d(aes(X, Y, fill=..level.., alpha=..level..), size=1, bins=128, data=subcrimes, geom="polygon") + 
        ggtitle(sprintf("Crime density by weight %d", i))
    ggsave(file=sprintf("weight_%s_density.png", i), scale=scale, width=width)
}

# distribution of crimes


# plot cluster hour 0, 1, on monday
repCrimes <- crimes[rep(1:nrow(crimes), crimes[,3]), 1:6]
weekday = weekdays[1+1][1]
i = 0
subcrimes0 = subset(repCrimes, Hour == i & DoW == 1)
clustered0 = kmeans(subcrimes0[5:6], 50, nstart=10) 
subcrimes0$Cluster = factor(clustered0$cluster)
centers0=as.data.frame(clustered0$centers)

ggmap(lamap) + 
    geom_point(data=subcrimes0, aes(X, Y, color=Cluster), alpha = 1/10) +
    geom_point(data=centers0, aes(X, Y), color="black") +
    guides(color=FALSE) +
    labs(title = sprintf("kmeans 50 clusters (%s @ %02d00)", weekday, i), y="Latitude", x="Longitude")

i = 1
subcrimes1 = subset(repCrimes, Hour == i & DoW == 1)
clustered1 = kmeans(subcrimes1[5:6], 50, nstart=10) 
subcrimes1$Cluster = factor(clustered1$cluster)
centers1=as.data.frame(clustered1$centers)

ggmap(lamap) + 
    geom_point(data=subcrimes1, aes(X, Y, color=Cluster), alpha = 1/10) +
    geom_point(data=centers1, aes(X, Y), color="black") +
    guides(color=FALSE) +
    labs(title = sprintf("kmeans 50 clusters (%s @ %02d00)", weekday, i), y="Latitude", x="Longitude")

 centers0$Weight <- table(subcrimes0$Cluster)
 centers1$Weight <- table(subcrimes1$Cluster)
 
 ggmap(lamap) + 
    geom_point(data=centers1, aes(X, Y, size=Weight), color="red") +
    guides(color=FALSE) +
    labs(title = sprintf("kmeans 50 cluster centroids (%s @ %02d00)", weekday, i), y="Latitude", x="Longitude")
    