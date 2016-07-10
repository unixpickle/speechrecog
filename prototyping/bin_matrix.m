function [mat] = bin_matrix(size)
  mat = zeros(size, size);
  for i = 1:(size/2+1)
    for j = 1:size
      mat(i, j) = cos(2*pi/size*(i-1)*(j-1));
    end
  end
  for i = 1:(size/2-1)
    for j = 1:size
      mat(i+size/2+1,j) = sin(2*pi/size*i*(j-1));
    end
  end
end
